package logger

import (
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

var Log, _ = zap.NewDevelopment()

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func Initialize(level string) error {
	config := zap.NewDevelopmentConfig()
	parsedLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	config.Level = parsedLevel

	Log, err = config.Build()
	if err != nil {
		return err
	}

	defer Log.Sync()
	return nil
}

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processingStart := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h(&lw, r)
		processingDuration := time.Since(processingStart)
		Log.Debug("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("processing duration", processingDuration.String()),
			zap.String("status code", strconv.Itoa(responseData.status)),
			zap.String("response size", strconv.Itoa(responseData.size)),
		)
	}
}
