FROM golang:1.21.2-bullseye

RUN mkdir /app
WORKDIR /app

ENV GOOS=linux GOARCH=amd64
COPY . .

# RUN go get
RUN go get 
RUN go build -o bin/gobank .


EXPOSE 3000 50051 9092

RUN chmod +x ./bin/gobank
# RUN #!/bin/ ./bin/gobank 

ENTRYPOINT [ "./bin/gobank" ]

