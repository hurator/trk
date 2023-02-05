package ice

type Trip struct {
	Trip struct {
		TripDate             string `json:"tripDate"`
		TrainType            string `json:"trainType"`
		Vzn                  string `json:"vzn"`
		ActualPosition       int    `json:"actualPosition"`
		DistanceFromLastStop int    `json:"distanceFromLastStop"`
		TotalDistance        int    `json:"totalDistance"`
		StopInfo             struct {
			ScheduledNext     string `json:"scheduledNext"`
			ActualNext        string `json:"actualNext"`
			ActualLast        string `json:"actualLast"`
			ActualLastStarted string `json:"actualLastStarted"`
			FinalStationName  string `json:"finalStationName"`
			FinalStationEvaNr string `json:"finalStationEvaNr"`
		} `json:"stopInfo"`
		Stops []struct {
			Station struct {
				EvaNr          string      `json:"evaNr"`
				Name           string      `json:"name"`
				Code           interface{} `json:"code"`
				Geocoordinates struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"geocoordinates"`
			} `json:"station"`
			Timetable struct {
				ScheduledArrivalTime    *int64 `json:"scheduledArrivalTime"`
				ActualArrivalTime       *int64 `json:"actualArrivalTime"`
				ShowActualArrivalTime   *bool  `json:"showActualArrivalTime"`
				ArrivalDelay            string `json:"arrivalDelay"`
				ScheduledDepartureTime  *int64 `json:"scheduledDepartureTime"`
				ActualDepartureTime     *int64 `json:"actualDepartureTime"`
				ShowActualDepartureTime *bool  `json:"showActualDepartureTime"`
				DepartureDelay          string `json:"departureDelay"`
			} `json:"timetable"`
			Track struct {
				Scheduled string `json:"scheduled"`
				Actual    string `json:"actual"`
			} `json:"track"`
			Info struct {
				Status            int    `json:"status"`
				Passed            bool   `json:"passed"`
				PositionStatus    string `json:"positionStatus"`
				Distance          int    `json:"distance"`
				DistanceFromStart int    `json:"distanceFromStart"`
			} `json:"info"`
			DelayReasons []struct {
				Code string `json:"code"`
				Text string `json:"text"`
			} `json:"delayReasons"`
		} `json:"stops"`
	} `json:"trip"`
	Connection struct {
		TrainType   interface{} `json:"trainType"`
		Vzn         interface{} `json:"vzn"`
		TrainNumber interface{} `json:"trainNumber"`
		Station     interface{} `json:"station"`
		Timetable   interface{} `json:"timetable"`
		Track       interface{} `json:"track"`
		Info        interface{} `json:"info"`
		Stops       interface{} `json:"stops"`
		Conflict    string      `json:"conflict"`
	} `json:"connection"`
	Active interface{} `json:"active"`
}

type Status struct {
	Connection   bool    `json:"connection"`
	ServiceLevel string  `json:"serviceLevel"`
	GpsStatus    string  `json:"gpsStatus"`
	Internet     string  `json:"internet"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	TileY        int     `json:"tileY"`
	TileX        int     `json:"tileX"`
	Series       string  `json:"series"`
	ServerTime   int64   `json:"serverTime"`
	Speed        float64 `json:"speed"`
	TrainType    string  `json:"trainType"`
	Tzn          string  `json:"tzn"`
	WagonClass   string  `json:"wagonClass"`
	Connectivity struct {
		CurrentState         string `json:"currentState"`
		NextState            string `json:"nextState"`
		RemainingTimeSeconds int    `json:"remainingTimeSeconds"`
	} `json:"connectivity"`
	BapInstalled bool `json:"bapInstalled"`
}
