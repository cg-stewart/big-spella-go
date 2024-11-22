package game

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

// MockDictionaryService is a mock implementation of DictionaryService
type MockDictionaryService struct {
	mock.Mock
}

func (m *MockDictionaryService) GetWordInfo(ctx context.Context, word string) (*Word, error) {
	args := m.Called(ctx, word)
	return args.Get(0).(*Word), args.Error(1)
}

func (m *MockDictionaryService) GenerateAudio(ctx context.Context, word string) ([]byte, error) {
	args := m.Called(ctx, word)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDictionaryService) GetHint(ctx context.Context, word *Word, hintType HintType) (string, error) {
	args := m.Called(ctx, word, hintType)
	return args.String(0), args.Error(1)
}

// MockWordService is a mock implementation of WordService
type MockWordService struct {
	mock.Mock
}

func (m *MockWordService) GetRandomWord(ctx context.Context, level int, category *string) (*Word, error) {
	args := m.Called(ctx, level, category)
	return args.Get(0).(*Word), args.Error(1)
}

func (m *MockWordService) ValidateSpelling(ctx context.Context, word, attempt string) bool {
	args := m.Called(ctx, word, attempt)
	return args.Bool(0)
}

func (m *MockWordService) TranscribeVoice(ctx context.Context, voiceData []byte) (string, error) {
	args := m.Called(ctx, voiceData)
	return args.String(0), args.Error(1)
}

// MockDB is a mock implementation of the database interface
type MockDB struct {
	*sqlx.DB
	mock.Mock
}

func NewMockDB() *MockDB {
	return &MockDB{
		DB: sqlx.NewDb(sql.OpenDB(mockConnector{}), "mock"),
	}
}

// mockConnector implements driver.Connector interface
type mockConnector struct{}

func (m mockConnector) Connect(context.Context) (sql.Conn, error) {
	return nil, nil
}

func (m mockConnector) Driver() driver.Driver {
	return nil
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	mockArgs := m.Called(ctx, query, args)
	if mockArgs.Get(0) == nil {
		return &sql.Row{}
	}
	return mockArgs.Get(0).(*sql.Row)
}

func (m *MockDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	mockArgs := m.Called(ctx, dest, query, args)
	return mockArgs.Error(0)
}

func (m *MockDB) CreateGame(ctx context.Context, game *Game) error {
	args := m.Called(ctx, game)
	return args.Error(0)
}

func (m *MockDB) GetGame(ctx context.Context, gameID string) (*Game, error) {
	args := m.Called(ctx, gameID)
	return args.Get(0).(*Game), args.Error(1)
}

func (m *MockDB) UpdateGame(ctx context.Context, game *Game) error {
	args := m.Called(ctx, game)
	return args.Error(0)
}
