dev-run-all:
	go run services/authentication/main.go & \
	go run services/booking/main.go & \
	go run services/notification/main.go & \
	go run services/driver_location/main.go


dev-run-frontend:
	cd frontend && npm run dev