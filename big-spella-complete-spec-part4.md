## API Endpoints & Handlers

```go
// api/http/routes.go

package http

import (
    "github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, services *Services) {
    // Auth routes
    auth := r.Group("/auth")
    {
        auth.POST("/register", handlers.Register(services.Auth))
        auth.POST("/login", handlers.Login(services.Auth))
        auth.POST("/refresh", handlers.RefreshToken(services.Auth))
    }

    // Game routes
    game := r.Group("/games").Use(middleware.Auth())
    {
        game.POST("/", handlers.CreateGame(services.Game))
        game.GET("/:id", handlers.GetGame(services.Game))
        game.POST("/:id/join", handlers.JoinGame(services.Game))
        game.POST("/:id/spell", handlers.SubmitSpelling(services.Game))
        game.GET("/active", handlers.ListActiveGames(services.Game))
    }

    // Tournament routes
    tournament := r.Group("/tournaments").Use(middleware.Auth())
    {
        tournament.POST("/", handlers.CreateTournament(services.Tournament))
        tournament.GET("/", handlers.ListTournaments(services.Tournament))
        tournament.GET("/:id", handlers.GetTournament(services.Tournament))
        tournament.POST("/:id/register", handlers.RegisterForTournament(services.Tournament))
    }

    // User routes
    users := r.Group("/users").Use(middleware.Auth())
    {
        users.GET("/me", handlers.GetProfile(services.User))
        users.PUT("/me", handlers.UpdateProfile(services.User))
        users.GET("/me/games", handlers.GetGameHistory(services.User))
        users.GET("/me/stats", handlers.GetStats(services.User))
    }

    // Premium routes
    premium := r.Group("/premium").Use(middleware.Auth())
    {
        premium.POST("/subscribe", handlers.Subscribe(services.Payment))
        premium.GET("/plans", handlers.ListPlans(services.Payment))
        premium.GET("/status", handlers.SubscriptionStatus(services.Payment))
    }
}

// Sample handler implementation
func CreateGame(gameService *game.GameService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input CreateGameInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(400, ErrorResponse{Error: err.Error()})
            return
        }

        userID := c.GetString("userID")
        game, err := gameService.CreateGame(c.Request.Context(), input, userID)
        if err != nil {
            c.JSON(500, ErrorResponse{Error: err.Error()})
            return
        }

        c.JSON(200, game)
    }
}
```

## WebSocket Implementation

```go
// api/ws/handler.go

package ws

import (
    "github.com/gorilla/websocket"
    "github.com/newrgm/bigspella/internal/game"
)

type Handler struct {
    upgrader  websocket.Upgrader
    game      *game.GameService
    stream    *stream.Client
    clients   map[string]*Client
    clientsMu sync.RWMutex
}

type Client struct {
    conn     *websocket.Conn
    userID   string
    gameID   string
    send     chan []byte
    handler  *Handler
}

func NewHandler(game *game.GameService, stream *stream.Client) *Handler {
    return &Handler{
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                // Implement origin checking
                return true
            },
        },
        game:    game,
        stream:  stream,
        clients: make(map[string]*Client),
    }
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("upgrade failed: %v", err)
        return
    }

    userID := r.Context().Value("userID").(string)
    gameID := r.URL.Query().Get("game_id")

    client := &Client{
        conn:    conn,
        userID:  userID,
        gameID:  gameID,
        send:    make(chan []byte, 256),
        handler: h,
    }

    h.registerClient(client)

    // Start client routines
    go client.writePump()
    go client.readPump()
}

func (c *Client) readPump() {
    defer func() {
        c.handler.unregisterClient(c)
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err,
                websocket.CloseGoingAway,
                websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }

        c.handleMessage(message)
    }
}

func (c *Client) handleMessage(message []byte) {
    var event WSEvent
    if err := json.Unmarshal(message, &event); err != nil {
        return
    }

    switch event.Type {
    case "spell":
        var attempt game.SpellingAttempt
        if err := json.Unmarshal(event.Payload, &attempt); err != nil {
            return
        }
        c.handler.game.HandleSpellingAttempt(context.Background(), attempt)

    case "join_game":
        var join JoinGameEvent
        if err := json.Unmarshal(event.Payload, &join); err != nil {
            return
        }
        c.handler.game.JoinGame(context.Background(), join.GameID, c.userID)
    }
}
```

## Notification System

```go
// internal/notification/service.go

package notification

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/service/ses"
)

type NotificationService struct {
    ses       *ses.Client
    templates *TemplateService
    queue     *Queue
}

type NotificationType string

const (
    NotificationGameInvite    NotificationType = "game_invite"
    NotificationTournamentStart NotificationType = "tournament_start"
    NotificationResults       NotificationType = "game_results"
    NotificationDailyWord     NotificationType = "daily_word"
)

type Notification struct {
    ID        string
    Type      NotificationType
    UserID    string
    Title     string
    Message   string
    Data      map[string]interface{}
    CreatedAt time.Time
}

func (s *NotificationService) SendNotification(ctx context.Context, notification *Notification) error {
    // Queue notification for processing
    if err := s.queue.Enqueue(ctx, notification); err != nil {
        return fmt.Errorf("enqueue notification: %w", err)
    }

    // Send email if user has email notifications enabled
    if s.shouldSendEmail(notification.Type, notification.UserID) {
        if err := s.sendEmail(ctx, notification); err != nil {
            log.Printf("failed to send email: %v", err)
        }
    }

    return nil
}

func (s *NotificationService) sendEmail(ctx context.Context, notification *Notification) error {
    template, err := s.templates.GetTemplate(notification.Type)
    if err != nil {
        return fmt.Errorf("get template: %w", err)
    }

    html, err := template.Execute(notification.Data)
    if err != nil {
        return fmt.Errorf("execute template: %w", err)
    }

    input := &ses.SendEmailInput{
        Destination: &ses.Destination{
            ToAddresses: []string{notification.Data["email"].(string)},
        },
        Message: &ses.Message{
            Body: &ses.Body{
                Html: &ses.Content{
                    Data: aws.String(html),
                },
            },
            Subject: &ses.Content{
                Data: aws.String(notification.Title),
            },
        },
        Source: aws.String("noreply@bigspella.com"),
    }

    _, err = s.ses.SendEmail(ctx, input)
    return err
}
```

## Error Handling & Monitoring

```go
// pkg/errors/errors.go

package errors

import (
    "fmt"
    "runtime"
)

type AppError struct {
    Code    string
    Message string
    Err     error
    Stack   []string
    Data    map[string]interface{}
}

func (e *AppError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code string, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Stack:   getStack(),
        Data:    make(map[string]interface{}),
    }
}

func Wrap(err error, code string, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
        Stack:   getStack(),
        Data:    make(map[string]interface{}),
    }
}

// pkg/monitoring/monitor.go

package monitoring

import (
    "context"
    "github.com/DataDog/datadog-go/statsd"
)

type Monitor struct {
    statsd *statsd.Client
    logger *Logger
}

func (m *Monitor) TrackGameMetrics(ctx context.Context, game *game.Game) {
    tags := []string{
        fmt.Sprintf("game_type:%s", game.Type),
        fmt.Sprintf("game_id:%s", game.ID),
    }

    m.statsd.Gauge("game.players", float64(len(game.Players)), tags, 1)
    m.statsd.Gauge("game.round", float64(game.Round), tags, 1)

    if game.Type == game.GameTypeRanked {
        m.statsd.Gauge("game.avg_elo", calculateAverageElo(game.Players), tags, 1)
    }
}

func (m *Monitor) TrackSpellingAttempt(ctx context.Context, attempt *game.SpellingAttempt, result *game.SpellingResult) {
    tags := []string{
        fmt.Sprintf("game_id:%s", attempt.GameID),
        fmt.Sprintf("correct:%t", result.Correct),
        fmt.Sprintf("input_type:%s", attempt.Type),
    }

    m.statsd.Increment("game.spelling_attempt", tags, 1)
    m.statsd.Timing("game.spelling_duration", 
        time.Since(attempt.StartTime),
        tags,
        1,
    )
}
```

Would you like me to proceed with the deployment configuration section next?