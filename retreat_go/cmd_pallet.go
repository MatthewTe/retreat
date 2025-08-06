package main

var APPLICATION_COMMANDS map[string]AppState = map[string]AppState{
	":feed":  FeedListState,
	":daily": DailyTaskTrackingState,
}

func ProcessCommands(fullCmd string) (rootCmd string, args string, err error) {

	// If its mood tracker - if no args query db for moods and render as timeseries.
	// if its mood tracker - if args, 2 args and one is add, next value is float/int and insert

	return "", "", nil
}
