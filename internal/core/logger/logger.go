package core_logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerContextKey struct{}

var (
	key = loggerContextKey{}
)

type Logger struct {
	*zap.Logger //без поля потому что встраиваем в структуру созданой структуры и все метожы этого логера будут доступны через через мою структуру это называется композиция ипа как наслежование

	file *os.File
}

func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(
		ctx,
		key,
		logger,
	)
}

func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(key).(*Logger)
	if !ok {
		panic("No logger in context")
	}
	return log
}
func NewLoger(config Config) (*Logger, error) {

	// ЭТАП 1: Парсинг и настройка уровня логирования
	// Создаем атомарный уровень. Он позволяет менять уровень логирования
	// (например, с INFO на DEBUG) прямо "на лету" во время работы приложения.
	zapLvl := zap.NewAtomicLevel()
	if err := zapLvl.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("Unmarshal log level: %w", err)
	}

	// ЭТАП 2: Подготовка папки для логов
	// Создаем папку, если её нет. Права 0755 (в восьмеричной системе) означают:
	// Владелец (Go-приложение) может всё (7), а группа и остальные — только читать и заходить (5).
	if err := os.MkdirAll(config.Folder, 0755); err != nil { // 0755 POSIX permissions =======>
		return nil, fmt.Errorf("Mkdir logFolder: %w", err)
	}

	// ЭТАП 3: Генерация уникального имени файла
	// Используем UTC время с точностью до микросекунд. Под большой нагрузкой
	// это гарантирует, что при перезапуске файлы логов не будут перезаписывать друг друга.
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000") // год месяц день часы минуты секунды вплоть до микросеркунд потому что за одну секунду в нащем проиложении может быть много хапросов и обрабоатыватся больщой трафик по этому мы сьавим столь сильную точность и так же что бы не было конфлита лог файлов х ха того что приложение под нашрухкой
	logFilePath := filepath.Join(
		config.Folder,
		fmt.Sprintf("%s.log", timestamp),
	)

	// ЭТАП 4: Открытие файла с флагами доступа
	// Побитовое ИЛИ (|) склеивает флаги: создать если нет (O_CREATE) + только для записи (O_WRONLY).
	// Права 0644: владелец читает/пишет (6), остальные только читают (4). Флаг выполнения (1) отсутствует.
	logfile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("Open log file: %w", err)
	}

	// ЭТАП 5: Настройка формата логов (Encoder)
	// Берем базовый конфиг для разработки (он делает логи понятными для человека, а не в JSON).
	zapConfig := zap.NewDevelopmentEncoderConfig()
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000")
	zapEncoder := zapcore.NewConsoleEncoder(zapConfig)

	// ЭТАП 6: Разветвление потоков (zapcore.NewTee)
	// NewTee работает как тройник: дублирует каждую запись лога одновременно
	// и в стандартный вывод терминала (os.Stdout), и в наш открытый файл на диске (logfile).
	core := zapcore.NewTee(
		zapcore.NewCore(zapEncoder, zapcore.AddSync(os.Stdout), zapLvl),
		zapcore.NewCore(zapEncoder, zapcore.AddSync(logfile), zapLvl),
	)

	// ЭТАП 7: Сборка финального логгера
	// zap.AddCaller() автоматически добавляет в лог имя файла и строку Go-кода, откуда пришел лог.
	zapLogger := zap.New(core, zap.AddCaller())

	// Возвращаем указатель на нашу структуру, упаковав туда сам логгер и файл.
	return &Logger{
		Logger: zapLogger,
		file:   logfile,
	}, nil
}
func (l *Logger) With(field ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(field...), file: l.file}
}
func (l *Logger) Close() {
	if err := l.file.Close(); err != nil {
		fmt.Println("Failed to close application logger:", err)
	}
}
