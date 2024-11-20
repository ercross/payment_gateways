package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"time"
)

var (
	ComponentKafka = Field{
		Key:   "Component",
		Value: "Kafka",
	}

	ComponentRedis = Field{
		Key:   "Component",
		Value: "Redis",
	}

	ComponentDatabase = Field{
		Key:   "Component",
		Value: "Database",
	}
)

// Level represents the severity of the log message.
type Level int8

const (
	DEBUG Level = iota - 1
	INFO
	WARN
	ERROR
	FATAL
)

type Config struct {
	Level       Level
	Destination io.WriteCloser
}

type Logger struct {
	zap    *zap.Logger
	config *Config
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func NewLogger(config *Config) (*Logger, error) {

	w := zapcore.AddSync(config.Destination)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		zapcore.Level(config.Level),
	)

	zapLogger := zap.New(core)
	return &Logger{zap: zapLogger, config: config}, nil
}

// NewSilentLogger creates a logger that silences all log output
func NewSilentLogger() (*Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.FatalLevel + 1)
	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{zap: zapLogger, config: &Config{}}, nil
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, toZapFields(fields)...)
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, toZapFields(fields)...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, toZapFields(fields)...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, toZapFields(fields)...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, toZapFields(fields)...)
}

func (l *Logger) Flush() error {
	return l.zap.Sync()
}

// RequestLogger is a middleware that logs HTTP requests
func RequestLogger(logger *Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrappedWriter, r)
			logger.Info("Request",
				NewField("method", r.Method),
				NewField("url", r.URL.String()),
				NewField("status", wrappedWriter.statusCode),
				NewField("client_ip", r.RemoteAddr),
				NewField("response_time", time.Since(start)),
			)
		})
	}
}

// toZapFields converts our custom Field type to zap.Field, keeping zap.Field encapsulated.
func toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// responseWriter is a custom http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code for logging
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
