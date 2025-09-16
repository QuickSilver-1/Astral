package logger

import (
	"log/slog"
	"os"
	"time"
)

func NewLogger(env string) *slog.Logger {
    var handler slog.Handler
    
    switch env {
    case "DEV":
        handler = setupDevelopmentHandler()
    case "PROD":
        handler = setupProductionHandler()
    default:
        handler = setupDefaultHandler()
    }
    
    return slog.New(handler)
}

func setupDevelopmentHandler() slog.Handler {
    return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level:       slog.LevelDebug,
        AddSource:   true, // Показывать файл и строку
        ReplaceAttr: developmentReplaceAttr,
    })
}

func setupProductionHandler() slog.Handler {
    return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:     slog.LevelInfo,
        AddSource: true,
    })
}

func setupDefaultHandler() slog.Handler {
    return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })
}

func developmentReplaceAttr(groups []string, a slog.Attr) slog.Attr {
    switch a.Key {
    case slog.TimeKey:
        if t, ok := a.Value.Any().(time.Time); ok {
            a.Value = slog.StringValue(t.Format("15:04:05.000"))
        }
    case slog.LevelKey:
        if level, ok := a.Value.Any().(slog.Level); ok {
            a.Value = slog.StringValue(colorizeLevel(level))
        }
    case slog.SourceKey:
        if source, ok := a.Value.Any().(*slog.Source); ok {
            source.File = shortenPath(source.File)
            a.Value = slog.AnyValue(source)
        }
    }
    return a
}

func colorizeLevel(level slog.Level) string {
    switch level {
    case slog.LevelDebug:
        return "\033[36mDEBUG\033[0m" // Cyan
    case slog.LevelInfo:
        return "\033[32mINFO\033[0m"  // Green
    case slog.LevelWarn:
        return "\033[33mWARN\033[0m"  // Yellow
    case slog.LevelError:
        return "\033[31mERROR\033[0m" // Red
    default:
        return level.String()
    }
}

func shortenPath(path string) string {
    return path
}