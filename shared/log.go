package shared

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/go-kit/kit/log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const (
	LvlDebug = "DEBUG"
	LvlInfo  = "INFO"
	LvlWarn  = "WARNING"
	LvlErr   = "ERROR"
)

func NewLogger(component string) *Logger {
	var kitlogger log.Logger
	kitlogger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	kitlogger = log.With(kitlogger, "ts", log.DefaultTimestampUTC)
	kitlogger = log.With(kitlogger, "component", component)

	return &Logger{
		kitlogger,
	}
}

type Logger struct {
	log.Logger
}

func (l *Logger) Debug(ctx context.Context, message string, keyvals ...interface{}) {
	l.logWithLvl(ctx, LvlDebug, message, keyvals)
}

func (l *Logger) Info(ctx context.Context, message string, keyvals ...interface{}) {
	l.logWithLvl(ctx, LvlInfo, message, keyvals)
}

func (l *Logger) Warn(ctx context.Context, message string, keyvals ...interface{}) {
	l.logWithLvl(ctx, LvlWarn, message, keyvals)
}

func (l *Logger) Err(ctx context.Context, message string, keyvals ...interface{}) {
	l.logWithLvl(ctx, LvlErr, message, keyvals)
}

// re-implement gorm logger
func (l *Logger) Print(v ...interface{}) {
	if len(v) > 1 {
		level := v[0]
		/*currentTime := "\n\033[33m[" + time.Now().Format("2006-01-02 15:04:05") + "]\033[0m"
		source := fmt.Sprintf("\033[35m(%v)\033[0m", v[1])
		keyvals := []interface{}{source, currentTime}*/
		keyvals := []interface{}{}

		if level == "sql" {
			keyvals = append(keyvals, "duration", fmt.Sprintf("%.2f", float64(v[2].(time.Duration).Nanoseconds()/1e4)/100.0))

			// sql
			var sql string
			var formattedValues []string

			for _, value := range v[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format(time.RFC3339)))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			}

			var formattedValuesLength = len(formattedValues)
			for index, value := range sqlRegexp.Split(v[3].(string), -1) {
				sql += value
				if index < formattedValuesLength {
					sql += formattedValues[index]
				}
			}

			keyvals = append(keyvals, "query", sql)
		} else {
			keyvals = append(keyvals, v[2:]...)
		}
		l.logWithLvl(context.Background(), LvlInfo, "new database query", keyvals)
	}
}

func (l *Logger) logWithLvl(ctx context.Context, lvl string, message string, keyvals []interface{}) {
	claims := ctx.Value("claims")
	roles := make([]string, 0)

	if claims != nil {
		for k, v := range claims.(map[string]interface{}) {
			switch k {
			case ROLE_ADMIN:
				if v.(bool) {
					roles = append(roles, ROLE_ADMIN)
				}
			case ROLE_TEACHER:
				if v.(bool) {
					roles = append(roles, ROLE_TEACHER)
				}
			case ROLE_ADULT:
				if v.(bool) {
					roles = append(roles, ROLE_ADULT)
				}
			case ROLE_OFFICE_MANAGER:
				if v.(bool) {
					roles = append(roles, ROLE_OFFICE_MANAGER)
				}
			}
		}
		keyvals = append(keyvals, "role", strings.Join(roles, "/"))
	}
	keyvals = append(keyvals, "level", lvl, "msg", message)
	l.Log(keyvals...)
}

var (
	sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)
)

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func (l *Logger) RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		l.Info(req.Context(), "new http request", "method", req.Method, "uri", req.RequestURI)

		next.ServeHTTP(w, req)
	})
}
