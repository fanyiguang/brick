package pipeline

const (
	modeLimitReport    = iota // limit触发
	modeIntervalReport        // 周期性触发
)

type Option[V any] func(line *PipeLine[V])

func SetLimit[V any](limit int) Option[V] {
	return func(line *PipeLine[V]) {
		if line.limit > 0 {
			line.limit = limit
		}
	}
}

func LimitReportMode[V any]() Option[V] {
	return func(line *PipeLine[V]) {
		line.mode = modeLimitReport
	}
}

func IntervalReportMode[V any]() Option[V] {
	return func(line *PipeLine[V]) {
		line.mode = modeIntervalReport
	}
}

func SetInterval[V any](interval int) Option[V] {
	return func(line *PipeLine[V]) {
		line.doInterval = interval
	}
}
