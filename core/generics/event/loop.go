package event

import (
	"github.com/izuc/zipp.foundation/core/workerpool"
)

var Loop *workerpool.UnboundedWorkerPool

func init() {
	Loop = workerpool.NewUnboundedWorkerPool("event.Loop")
	Loop.Start()
}
