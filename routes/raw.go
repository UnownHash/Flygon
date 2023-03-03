package routes

type RawBody struct {
	Uuid         string      `json:"uuid" binding:"required"`
	Username     string      `json:"username" binding:"required"`
	HaveAr       bool        `json:"have_ar"`
	TrainerExp   int         `json:"trainerexp" default:"0"`
	TrainerLevel int         `json:"trainerLevel" default:"0"`
	TrainerLvl   int         `json:"trainerlvl" default:"0"`
	Contents     interface{} `json:"contents"` // only one of those three is needed
	Protos       interface{} `json:"protos"`   // only one of those three is needed
	GMO          interface{} `json:"gmo"`      // only one of those three is needed
}

func Raw(body RawBody) (interface{}, error) {
	return nil, nil
}
