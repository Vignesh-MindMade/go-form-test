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



for env file need this import

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


For API

http://localhost:8080/api/users

with body-> form-data
| Key   | Type     | Value                                   |
| ----- | -------- | --------------------------------------- |
| name  | Text     | Vignesh                                 |
| email | Text     | [test@gmail.com](mailto:test@gmail.com) |
| phone | Text     | 9876543210                              |
| city  | Text     | Chennai                                 |
| image | File     | select image.jpg                        |
| pdf   | File     | select resume.pdf                       |


Upload API limits

Max request size: 200 MB

Format: multipart/form-data

Fields: name, email, phone, city

Files: image (jpg/png), pdf

Error: 400 if payload exceeds limit