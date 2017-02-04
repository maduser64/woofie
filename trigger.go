// Network-triggered randomized sound player, simulating how a dog would bark at
// a door.

// This file specifies the WoofTrigger interface, which HTTP, UDP, and any
// future broadcasts will plug into.

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package woofie

import (
	"log"
)

type WoofTrigger interface {
	MainLoop(logger *log.Logger, woofer *Woofer) error
}
