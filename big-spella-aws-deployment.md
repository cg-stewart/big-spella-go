# Big Spella - AWS Deployment Configuration

## Infrastructure Overview

```hcl
# terraform/main.tf

provider "aws" {
  region = "us-east-1"
}

# VPC and Network Configuration
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "bigspella-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true  # Cost optimization for dev/staging
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "bigspella-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ECS Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = "bigspella"
  requires_compatibilities = ["FARGATE"]
  network_mode            = "awsvpc"
  cpu                     = "512"
  memory                  = "1024"
  execution_role_arn      = aws_iam_role.ecs_execution.arn
  task_role_arn           = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "bigspella"
      image = "${aws_ecr_repository.app.repository_url}:latest"
      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "APP_ENV"
          value = "production"
        },
        {
          name  = "REDIS_URL"
          value = aws_elasticache_cluster.redis.cache_nodes[0].address
        }
      ]
      secrets = [
        {
          name      = "DB_CONNECTION_STRING"
          valueFrom = aws_secretsmanager_secret.db_url.arn
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = "/ecs/bigspella"
          awslogs-region        = "us-east-1"
          awslogs-stream-prefix = "ecs"
        }
      }
    }
  ])
}

# ECS Service
resource "aws_ecs_service" "app" {
  name            = "bigspella"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  launch_type     = "FARGATE"
  desired_count   = 2

  network_configuration {
    subnets         = module.vpc.private_subnets
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "bigspella"
    container_port   = 8080
  }
}

# Application Load Balancer
resource "aws_lb" "app" {
  name               = "bigspella-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets           = module.vpc.public_subnets
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.app.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

# RDS Database
resource "aws_db_instance" "postgres" {
  identifier        = "bigspella-db"
  engine           = "postgres"
  engine_version   = "15.4"
  instance_class   = "db.t4g.micro"  # Start small
  allocated_storage = 20

  db_name  = "bigspella"
  username = "bigspella"
  password = random_password.db_password.result

  vpc_security_group_ids = [aws_security_group.db.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name

  backup_retention_period = 7
  skip_final_snapshot    = true  # For dev/staging
}

# Redis for Game State
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "bigspella-redis"
  engine              = "redis"
  node_type           = "cache.t4g.micro"  # Start small
  num_cache_nodes     = 1
  port                = 6379
  security_group_ids  = [aws_security_group.redis.id]
  subnet_group_name   = aws_elasticache_subnet_group.main.name
}
```

## Docker Configuration

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git build-base

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o bigspella ./cmd/server

# Final stage
FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bigspella .
COPY config ./config

ENV GIN_MODE=release
ENV TZ=UTC

EXPOSE 8080

ENTRYPOINT ["/app/bigspella"]
```

## GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy to AWS

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

env:
  AWS_REGION: us-east-1
  ECR_REPOSITORY: bigspella
  ECS_CLUSTER: bigspella-cluster
  ECS_SERVICE: bigspella
  TASK_DEFINITION: bigspella

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        
    - name: Run tests
      run: go test -v ./...

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: ${{ env.AWS_REGION }}
    
    - name: Login to Amazon ECR
      id: login-ecr
      uses: aws-actions/amazon-ecr-login@v1
    
    - name: Build and push image to ECR
      id: build-image
      env:
        ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
        IMAGE_TAG: ${{ github.sha }}
      run: |
        docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG .
        docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
        echo "::set-output name=image::$ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG"
    
    - name: Update ECS service
      run: |
        aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --force-new-deployment
```

## Application Configuration

```go
// config/config.go
package config

type Config struct {
    Port     string `envconfig:"PORT" default:"8080"`
    Database struct {
        URL string `envconfig:"DB_CONNECTION_STRING" required:"true"`
    }
    Redis struct {
        URL string `envconfig:"REDIS_URL" required:"true"`
    }
    AWS struct {
        Region          string `envconfig:"AWS_REGION" default:"us-east-1"`
        ChimeMediaRegion string `envconfig:"AWS_CHIME_MEDIA_REGION" default:"us-east-1"`
    }
    OpenAI struct {
        Key string `envconfig:"OPENAI_API_KEY" required:"true"`
    }
    GetStream struct {
        Key    string `envconfig:"GETSTREAM_API_KEY" required:"true"`
        Secret string `envconfig:"GETSTREAM_API_SECRET" required:"true"`
    }
}
```

## Security Groups

```hcl
# terraform/security.tf

resource "aws_security_group" "alb" {
  name        = "bigspella-alb"
  description = "ALB security group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    protocol    = "tcp"
    from_port   = 443
    to_port     = 443
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "ecs_tasks" {
  name        = "bigspella-ecs-tasks"
  description = "ECS tasks security group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    protocol        = "tcp"
    from_port       = 8080
    to_port         = 8080
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

This configuration:
1. Uses ECS Fargate for container orchestration
2. Starts with small instance sizes to minimize costs
3. Implements auto-scaling based on demand
4. Sets up CI/CD with GitHub Actions
5. Uses AWS ECR for container registry
6. Configures ALB for load balancing
7. Implements proper security groups and VPC setup

Would you like me to:
1. Add monitoring and logging configurations?
2. Include backup strategies?
3. Add cost optimization recommendations?
4. Provide scaling configurations?