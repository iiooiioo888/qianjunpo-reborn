package lockstep

import "time"

// LockstepTurnFPS is the authoritative simulation turn rate (10 fps).
const LockstepTurnFPS = 10

// LockstepTurnDuration is wall-clock duration of one lockstep turn.
const LockstepTurnDuration = time.Millisecond * 100

// CommandDelayFrames is how many lockstep frames before a command executes.
const CommandDelayFrames = 3

// GameTurnFrames is the number of logic sub-frames inside each lockstep turn.
const GameTurnFrames = 6
