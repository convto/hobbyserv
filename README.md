# hobbyserv
Hobby server for pprof test

## running server
```
$ go run main.go
```

## create user
```
 curl -X POST localhost:9999/users/create -d '{"email":"example@convto.com","password":"123456"}'
```

## login user
```
 curl -X POST localhost:9999/users/create -d '{"email":"example@convto.com","password":"123456"}'
```