# go-framework
This is framework in go or golang. It is designed to allow a developer to create and start a Rest Api server or a load balancer under a few seconds.

The data stores that this framework can connect to are:
1. Any flavour of SQL (Example: SQL, MicrosoftSQL, MariaDB, Postgres, YSQL, ...)
2. Any flavour of Cassandra (Example: YCQL,...)
3. Mongo DB
4. Redis

Note: Using this framework we can connect to 4 data stores simultaneously i.e. 1 flavour of SQL, 1 flavour of CQL, MongoDB and Redis.

Instructions to start a <b>SERVER</b>:
1. Initialize you tools
2. Initialize a new router
3. Setup your endpoints
4. Connect to the data stores (Optional)
5. Start your server

NOTE: a server can be started without connecting to a DB, i.e. using t.ConnectToDB() is optional.

<b>*Example*</b>
Inside `main.go`
```
func main() {
	var t = tools.NewTools()
	t.MyRouter = tools.NewMyRouter()
	t.NewRestRequest("GET", "", nil)
	t.NewRestRequest("DELETE", "", nil)
	t.NewRestRequest("PUT", "", nil)
	t.NewRestRequest("POST", "", defaultPostPage) // also can use t.NewRestRequest("GET", "", nil)
	t.ConnectToDB() // optional
	t.Start("")
}

func defaultPostPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "default Page for POST!")
	fmt.Println("Endpoint Hit: default Page")
}
```

Instructions to  <b>connect to a DB</b>:
1. Type all your data base configurations inside `config/app.env`
2. Use `t.ConnectToDB()` in your program

<b>*Example*</b>

Inside `config/app.env`
```
SQL_DRIVER_NAME=postgres
SQL_HOST=localhost
SQL_PORT=5432
SQL_USER=postgres
SQL_PASSWORD=password
SQL_DB_NAME=postgres
...
```

Instructions to start a <b>LOAD BALANCER</b> are simple and given below:
1. Type the adresses of all servers over which you would like to balance the load in `config/loadBalancerApp.env`
2. start the load balancer

<b>*Example*</b>

Inside `config/loadBalancerApp.env`
```
LOAD_BALANCER_HOSTS=http://127.0.0.1:10000,http://127.0.0.1:10001
```

Inside `main.go`
```
var t = tools.NewTools()
t.StartLoandBalancer()
```
