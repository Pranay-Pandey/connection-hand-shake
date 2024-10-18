package models

type AdminUser struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Driver struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

type Admin struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

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
