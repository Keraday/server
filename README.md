- Enviroment:
- ENV=dev || prod
- LOG_LVL=debug || info || warn || error
- SERV_ADDR=localhost:8080
- DB_URL=postgres://[login]:[password]@localhost:5432/postgres?sslmode=disable

 
 
 
- migrate -source file://. -database postgres://postgres:edifier@localhost:5432/postgres?sslmode=disable up
