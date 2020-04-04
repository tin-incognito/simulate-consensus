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
	BACKVIEWCHANGE = "back-viewchange"
	BACKNEWVIEW = "back-newview"
	VERIFYNEWVIEW = "verify-newview"
	BACKVERIFYNEWVIEW = "back-verify-newview"
	VIEWCHANGEFINISH = "viewchange-finish"
)

//Faulty mode
const (
	FAULTY = "faulty"

)

//Mode
const (
	NormalMode = "normal_mode"
	ViewChangeMode = "viewchange_mode"
)

//Viewchange flow:
// viewchange -> newview -> normal mode

//Normal flow:
//broadcast -> pre prepare -> prepare -> commit

//Faulty flow:
//sub-phase -> viewchange -> ... -> end of viewchange flow -> normal mode