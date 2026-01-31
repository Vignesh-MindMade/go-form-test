Installed go using "choco"

mkdir goform
cd goform

To run the progrom
    > go main.go
    > go build -> to create executable

Open in editor
step-1

    > go mod init goform

step-2
    > create main.go file

create a Db i am suing mysql

create a table name users with id,name and email

to talk with db go must have this deiver of db

step-3
    > go get -u github.com/go-sql-driver/mysql



for env file

go get github.com/joho/godotenv






For query for all image and pdf
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    city VARCHAR(50),
    image_path VARCHAR(255),
    pdf_path VARCHAR(255)
);