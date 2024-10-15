package models

type User struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VehicleDriver struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	VehicleID     string `json:"vehicleID"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	VehicleType   string `json:"vehicleType"`
	VehicleVolume string `json:"vehicleVolume"`
}

type Booking struct {
	ID          int32  `json:"id"`
	UserID      string `json:"userID"`
	VehicleID   string `json:"vehicleID"`
	Price       string `json:"price"`
	Pickup      string `json:"pickup"`
	Dropoff     string `json:"dropoff"`
	BookedAt    string `json:"bookedAt"`
	CompletedAt string `json:"completedAt"`
}

type AdminUser struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
