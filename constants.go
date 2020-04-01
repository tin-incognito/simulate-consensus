package main

const (
	VIEWCHANGE = "viewchange"
	PREPREPARE = "preprepare"
	PREPARE = "prepare"
	COMMIT = "commit"
	PREPAREVIEWCHANGE = "prepare-viewchange"
	NEWVIEW = "newview"
	BACKVIEWCHANGE = "back-viewchange"
	BACKNEWVIEW = "back-newview"
	VERIFYNEWVIEW = "verify-newview"
	BACKVERIFYNEWVIEW = "back-verify-newview"
	VIEWCHANGEFINISH = "viewchange-finish"
)

const (
	NormalMode = "normal_mode"
	ViewChangeMode = "viewchange_mode"
)

//prepareviewchange -> viewchange -> backviewchange -> newview -> verify-newview -> back-verify-newview -> back-newview -> viewchange finish -> normal mode