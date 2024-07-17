This is a To Do api. You can create an postgress docker an keep database there.
# Quick Start
```
docker run --name todoapipg -e -p 5432:5432 POSTGRES_PASSWORD=mysecretpassword -d postgres
go mod tidy
make run
```
* create an postgres docker container and bind port 5432 of container to 5432 of your host machine
* download necessary module (which should not more than few)
* build and run the project


### After login get the JWT and put it in the header with key "jwt-token"

```
{
    "jwt-token": [token]
}
```

# Actions 

## /account

### GET: returns all accounts (this is kind of for admins)

### POST: creates a new account
```
// example request body (json):
{
    "first_name": "Umut",
    "last_name": "Ozkan",
    "password": "supersecretpassword"
}
```

## /account/{id}
example /account/3

### GET: return all todos belonging to account with specified id if JWT is valid

### POST: modify (add, change status(done/ not done), delete) To Dos
```
// example request body (json):
{
    // list of todos to add
    "add": [
        {
            status = false,
            context = "Fix delete account bugs"
        },
        ...
    ], 
    // list of todos to change status 
    // number: number (something like a todo id) of this todo belonging to account
    // done: is goal done or not
    "status_change" : [
        {
            "number": 2, 
            "done": true 
        }, 
        ...
    ],
    // list of todo numbers to delete
    "delete": [
        4,
        5,
        6
    ]
}
```

### DELETE: delete the account if password is correct and JWT is valid
```
// example request body (json):
{
    "password": "supersecretpassord"
}
```

## /login/{id}
example /login/15

### POST: get the password compare, create and give JWT
```
// example request body (json):
{
    "password": "supersecretpassord"
}
```