package stream

import (
	"desktop-cleaner/internal"
	streamtui "desktop-cleaner/stream_tui"
	"desktop-cleaner/types"
	"log"
)

var OnStreamPlan types.OnStreamPlan = func(params types.OnStreamPlanParams) {
	if params.Err != nil {
		log.Println("Error in stream:", params.Err)
		return
	}

	if params.Msg.Type == internal.StreamMessageStart {
		log.Println("Stream started")
		return
	}

	// log.Println("Stream message:")
	// log.Println(spew.Sdump(*params.Msg))

	streamtui.Send(*params.Msg)
}
