# Yu-Go-Oh-API

# Running the backend
  First run the docker-compose
  > docker-compose up 
  
  Run the go project
  > go run main.go
  
  The API defaults to localhost on port 4000.
 
 # Preparations
  You will need an JSON containing information of all the cards. 
  After obtaining it, edit dbUtils.go at `gopostgres/dbutils/dbUtils.go`
  changing line 186 to your json.
  Now, go to `/cards/load` on the API and it will auto load every card on JSON.
  
  Everything finished, you're all set. Enjoy the API!
