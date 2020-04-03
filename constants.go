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
	FaultyMode = "faulty-mode"
	AfterFaultyMode = "after-faulty-mode"
)

//Viewchange flow:
// viewchange -> new

//Normal flow:
//broadcast -> pre prepare -> prepare -> commit