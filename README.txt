-inspired by - https://medium.com/@matryer/production-ready-mongodb-in-go-for-beginners-ef6717a77219
-just pure mongo db and go 
- see how to use Adapters
- we use withDB to get an adapter 
- which is then used to decorate handlers which gives them access to db session 
- it is done by copying a database session in a context and using the keyword database


to test it 
- mongod --dbpath=./db
- go run main.go
- curl -i -X POST -d '{"name":"mazda","description":"car","floor":1,"unit":105}' http://localhost:8080/companies
