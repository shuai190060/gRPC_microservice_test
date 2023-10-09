FROM golang:1.21.2-bullseye

RUN mkdir /app
WORKDIR /app
ENV GOPATH /app

COPY . .

# RUN go get
RUN go build -o bin/gobank .


EXPOSE 3000 50051

RUN chmod +x ./bin/gobank
# RUN #!/bin/ ./bin/gobank 

ENTRYPOINT [ "./bin/gobank" ]

