# url-checker
This URL checker service is used to check whether the input URL is safe or has malicious resources.

### Instructions to Run

1) Prerequisites – Make sure to install – Git, Docker (which comes with docker-compose), Golang, python3, Make, Curl or Postman (for API testing)
2) The init-db.js has already been populated with the relevant insert commands, so there is no need to run the python load script.
3) From the project root run “make up”. This will instantiate the application, database, and cache containers. Wait for a few seconds until you see “MongoDB init process complete; ready for startup”. Since mongoDB is inserting about 150,000 records it take some time for the init process to complete.
4) Do the API testing by running curl commands – Refer to examples section in the final report.
5) After the API testing is done, press cmd + c to exit, and then issue “make clear” to remove all containers.
6) If you don’t have make installed then you need to run “docker-compose build” and “docker-compose up” to run the server, and then “docker-compose down” to remove the containers.
