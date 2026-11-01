package core_http_middleware

import (
	"net/http"
	core_logger "notification/internal/core/logger"
	core_http_response "notification/internal/core/transport/http/response"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIdHeader = "X-Request-ID"

// 1. МИДДЛВАР: RequestID
// ДЛЯ ЧЕГО: Это фундамент цепочки. Его единственная задача — гарантировать,
// что у КАЖДОГО входящего запроса будет свой уникальный след (UUID).
// Без этого миддлвара мы не сможем связать логи и отследить конкретный клик пользователя.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пытаемся достать ID из заголовков (если его прислал фронтенд или Гейтвей)
			requestID := r.Header.Get(requestIdHeader)
			// Если никто ID не прислал, генерируем новый UUID самостоятельно
			if requestID == "" {
				requestID = uuid.NewString()
			}
			// Записываем ID обратно в заголовки запроса (чтобы следующие миддлвары могли его прочитать)
			r.Header.Set(requestIdHeader, requestID)
			// Отдаем этот же ID в ответе клиенту (чтобы в случае ошибки юзер мог назвать этот ID поддержке)
			w.Header().Set(requestIdHeader, requestID)
			// Передаем управление следующему слою
			next.ServeHTTP(w, r)
		})
	}
}

// 2. МИДДЛВАР: Logger
// ДЛЯ ЧЕГО: Он "заряжает" логгер контекстом текущего запроса.
// Вместо того чтобы каждый раз вручную писать zap.String("request_id", ...),
// этот миддлвар создает копию логгера, в которую НАВСЕГДА вшиты RequestID и URL текущего запроса,
// и прокидывает этот логгер глубоко внутрь контекста для хендлеров и бизнес-логики.
func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Читаем RequestID, который на предыдущем этапе заботливо подготовил миддлвар RequestID
			requestID := r.Header.Get(requestIdHeader)
			// Метод .With() создает новый изолированный логгер, который к каждой строчке лога
			// будет АВТОМАТИЧЕСКИ приписывать request_id и вызываемый url.
			l := log.With(
				zap.String("request_id", requestID),
				zap.String("url", r.URL.String()),
			)
			// Кладем этот кастомизированный логгер в контекст запроса под ключом "log"
			ctx := core_logger.ToContext(r.Context(), l)
			// Передаем запрос дальше, но уже с ОБНОВЛЕННЫМ контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// 3. МИДДЛВАР: Trace
// ДЛЯ ЧЕГО: Он замеряет производительность (Latency) и логирует жизненный цикл запроса.
// Ты видишь в логах две четкие точки: когда запрос вошел в систему, и когда он полностью завершился,
// а также с каким HTTP-статусом он ушел клиенту и сколько миллисекунд это заняло.
func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Достаем логгер из контекста (он уже содержит request_id благодаря миддлвару Logger)
			log := core_logger.FromContext(ctx)
			// Оборачиваем стандартный ResponseWriter в обертку,
			// которая умеет перехватывать и запоминать HTTP статус-код (например, 200, 400, 500).
			rw := core_http_response.NewResponseWritter(w)
			// Фиксируем точное время старта запроса
			before := time.Now()
			// Пишем дебаг-лог, что запрос пошел в обработку
			log.Debug(
				">>> Incoming HTTP request",
				zap.String("HTTPMethod", r.Method),
				zap.Time("Time", before.UTC()),
			)
			// Передаем обертку rw вместо стандартного w дальше в хендлеры,
			// чтобы когда хендлер вызовет w.WriteHeader(200), наша обертка это записала.
			next.ServeHTTP(rw, r)
			// К этому моменту хендлер ОТРАБОТАЛ. Запрос полностью выполнен.
			// Считаем сколько времени прошло: time.Since(before)
			log.Debug(
				">>> Done HTTP request <<<",
				zap.Int("status_code", rw.GetStatusCode()),
				zap.Duration("Latency", time.Since(before)),
			)
		})
	}
}

// 4. МИДДЛВАР: Panic
// ДЛЯ ЧЕГО: Это твой подушка безопасности (страховка).
// Если внутри твоего хендлера или доменной логики случится критическая ошибка
// (например, обращение к nil-указателю, из-за чего Go вызовет panic),
// этот миддлвар поймает её с помощью recover().
// Он не даст приложению упасть (Crash), красиво залогирует панику и вернет клиенту аккуратный 500-й ответ.
func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Достаем из контекста логгер, который на предыдущем этапе туда положил миддлвар Logger
			log := core_logger.FromContext(ctx)
			// Инициализируем твой кастомный хелпер для ответов, передавая ему логгер и ответчик
			responseHandler := core_http_response.NewHttpResponseHandler(log, w)
			// defer выполнится ВСЕГДА, даже если код ниже (next.ServeHTTP) взорвется паникой
			defer func() {
				if p := recover(); p != nil {
					// Поймали панику! responseHandler внутри себя сделает log.Error() со всем стек-трейсом
					// и отправит клиенту JSON-ответ об ошибке, сохранив микросервису жизнь.
					responseHandler.PanicResponse(p, "During handle http request got unexpected panic ")
				}
			}()
			// Пропускаем запрос дальше вглубь приложения
			next.ServeHTTP(w, r)
		})
	}
}

func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigins := map[string]struct{}{
				"http://localhost:5050": {},
			}
			origin := r.Header.Get("Origin")
			if _, ok := allowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
