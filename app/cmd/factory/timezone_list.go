package factory

import (
	"github.com/mondegor/go-core/util/timezone"

	"print-shop-back/config"
	"print-shop-back/internal/adapter/log"
)

// InitTimeZones - создаёт объект timezone.LocationList.
// Список готовится один раз при старте: часовые пояса читаются с диска,
// поэтому обращений к базе часовых поясов в процессе обработки запросов
// уже не происходит. UTC и Local регистрируются всегда и в конфиге не нужны.
//
// Ошибку не возвращает: NewLocationList негодные имена молча пропускает,
// а отвергаются они раньше, при загрузке конфигурации (config.ValidateStorableTimeZones).
func InitTimeZones(logger log.Logger, cfg config.Config) *timezone.LocationList {
	log.Info(logger, "Create and init timezone list")

	locations := timezone.NewLocationList(cfg.AppTimeZones)

	log.DebugFunc(
		logger,
		func() string {
			var buf []byte

			buf = append(buf, "Time zones:"...)

			// имена берутся из конфигурации, а не у списка: он их не публикует,
			// а негодные из них отвергнуты раньше, при загрузке конфигурации
			for _, name := range cfg.AppTimeZones {
				buf = append(buf, "\n- "+name...)
			}

			return string(buf)
		},
	)

	return locations
}
