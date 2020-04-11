package main

//Normal mode
const (
	PREPREPARE = "preprepare"
	PREPARE = "prepare"
	COMMIT = "commit"
)

//Viewchange mode
const (
	VIEWCHANGE = "viewchange"
	PREPAREVIEWCHANGE = "prepare-viewchange"
	NEWVIEW = "newview"
)

//Mode
const (
	NormalMode = "normal_mode"
	ViewChangeMode = "viewchange_mode"
)

const (
	PREPARETIMER = "prepare-timer"
	COMMITTIMER = "commit-timer"
)

const (
	Production = "prod"
	Test = "test"
)

//Viewchange flow:
// viewchange -> newview -> normal mode

//Normal flow:
//broadcast -> pre prepare -> prepare -> commit

//Faulty flow:
//sub-phase -> viewchange -> ... -> end of viewchange flow -> normal mode