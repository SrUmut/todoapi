package main

func main() {
	db, err := NewPostgresStore()
	if err != nil {
		panic(err)
	}
	if err := db.Init(); err != nil {
		panic(err)
	}

	api := newAPIServer(":8080", db)
	if err := api.Start(); err != nil {
		panic(err)
	}
}
