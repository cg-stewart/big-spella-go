## Monitoring Configuration

```hcl
# terraform/monitoring.tf

# CloudWatch Dashboard
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "bigspella-metrics"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/ECS", "CPUUtilization", "ServiceName", "bigspella", "ClusterName", "bigspella-cluster"],
            [".", "MemoryUtilization", ".", ".", ".", "."]
          ]
          period = 300
          region = "us-east-1"
          title  = "ECS Service CPU & Memory"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/ApplicationELB", "RequestCount", "LoadBalancer", aws_lb.app.name],
            [".", "TargetResponseTime", ".", "."]
          ]
          period = 300
          region = "us-east-1"
          title  = "ALB Metrics"
        }
      }
    ]
  })
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "service_cpu" {
  alarm_name          = "bigspella-cpu-utilization"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"

  dimensions = {
    ClusterName = aws_ecs_cluster.main.name
    ServiceName = aws_ecs_service.app.name
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "db_storage" {
  alarm_name          = "bigspella-db-storage"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "FreeStorageSpace"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = "5000000000" # 5GB

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.postgres.id
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
}

# SNS Topic for Alerts
resource "aws_sns_topic" "alerts" {
  name = "bigspella-alerts"
}

resource "aws_sns_topic_subscription" "alerts_email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = "alerts@bigspella.com"
}
```

## Logging Configuration

```go
// internal/logging/config.go
package logging

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type Logger struct {
    cloudwatch *cloudwatchlogs.Client
    logGroup   string
    logStream  string
}

func NewLogger(ctx context.Context) (*Logger, error) {
    client := cloudwatchlogs.NewFromConfig(awsConfig)
    
    logger := &Logger{
        cloudwatch: client,
        logGroup:   "/bigspella/app",
        logStream:  "app-" + time.Now().Format("2006-01-02"),
    }

    // Ensure log group exists
    err := logger.createLogGroupIfNotExists(ctx)
    if err != nil {
        return nil, err
    }

    return logger, nil
}

// terraform/logging.tf
resource "aws_cloudwatch_log_group" "app" {
  name              = "/bigspella/app"
  retention_in_days = 30  # Adjust based on needs
}

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/bigspella"
  retention_in_days = 30
}
```

## Backup Strategy

### Database Backups

```hcl
# terraform/backup.tf

# AWS Backup vault
resource "aws_backup_vault" "main" {
  name = "bigspella-backup-vault"
}

# AWS Backup plan
resource "aws_backup_plan" "main" {
  name = "bigspella-backup-plan"

  rule {
    rule_name         = "daily_backup"
    target_vault_name = aws_backup_vault.main.name
    schedule          = "cron(0 5 ? * * *)" # Daily at 5 AM UTC

    lifecycle {
      delete_after = 30 # Keep backups for 30 days
    }
  }

  rule {
    rule_name         = "weekly_backup"
    target_vault_name = aws_backup_vault.main.name
    schedule          = "cron(0 5 ? * 1 *)" # Weekly on Sunday at 5 AM UTC

    lifecycle {
      delete_after = 90 # Keep weekly backups for 90 days
    }
  }
}

# Select resources for backup
resource "aws_backup_selection" "main" {
  name         = "bigspella-backup-selection"
  plan_id      = aws_backup_plan.main.id
  iam_role_arn = aws_iam_role.backup.arn

  resources = [
    aws_db_instance.postgres.arn
  ]
}

# IAM role for AWS Backup
resource "aws_iam_role" "backup" {
  name = "bigspella-backup-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "backup.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "backup" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup"
  role       = aws_iam_role.backup.name
}
```

### Application Data Backup

```go
// internal/backup/service.go
package backup

import (
    "context"
    "time"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type BackupService struct {
    s3Client *s3.Client
    bucket   string
}

func (s *BackupService) BackupGameData(ctx context.Context, gameID string) error {
    // Export game data to JSON
    gameData, err := s.gameService.ExportGame(ctx, gameID)
    if err != nil {
        return fmt.Errorf("export game: %w", err)
    }

    // Upload to S3
    key := fmt.Sprintf("game-backups/%s/%s.json",
        time.Now().Format("2006-01-02"),
        gameID,
    )

    _, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
        Body:   bytes.NewReader(gameData),
    })

    return err
}

// terraform/s3.tf
resource "aws_s3_bucket" "backups" {
  bucket = "bigspella-backups"
}

resource "aws_s3_bucket_lifecycle_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    id     = "game_data_lifecycle"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = 365
    }
  }
}
```

## Custom Metrics and Monitoring

```go
// internal/metrics/service.go
package metrics

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

type MetricsService struct {
    cloudwatch *cloudwatch.Client
    namespace  string
}

func (s *MetricsService) TrackGameMetrics(ctx context.Context, game *Game) error {
    metrics := []cloudwatch.MetricDatum{
        {
            MetricName: aws.String("ActivePlayers"),
            Value:      aws.Float64(float64(len(game.Players))),
            Unit:       cloudwatch.StandardUnitCount,
        },
        {
            MetricName: aws.String("RoundDuration"),
            Value:      aws.Float64(game.LastRoundDuration.Seconds()),
            Unit:       cloudwatch.StandardUnitSeconds,
        },
        {
            MetricName: aws.String("CorrectAnswers"),
            Value:      aws.Float64(float64(game.CorrectAnswers)),
            Unit:       cloudwatch.StandardUnitCount,
        },
    }

    _, err := s.cloudwatch.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
        Namespace:  aws.String(s.namespace),
        MetricData: metrics,
    })

    return err
}

// Add custom dashboard for game metrics
resource "aws_cloudwatch_dashboard" "game_metrics" {
  dashboard_name = "bigspella-game-metrics"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["BigSpella/Games", "ActivePlayers", "GameType", "solo"],
            [".", ".", ".", "group"],
            [".", ".", ".", "tournament"]
          ]
          period = 300
          region = "us-east-1"
          title  = "Active Players by Game Type"
        }
      },
      {
        type   = "metric"
        width  = 12
        height = 6
        properties = {
          metrics = [
            ["BigSpella/Games", "CorrectAnswers"],
            [".", "TotalAttempts"]
          ]
          period = 300
          region = "us-east-1"
          title  = "Answer Statistics"
        }
      }
    ]
  })
}
```

Would you like me to:
1. Add error tracking configuration?
2. Include performance monitoring specifics?
3. Add disaster recovery procedures?
4. Detail specific monitoring use cases for game events?