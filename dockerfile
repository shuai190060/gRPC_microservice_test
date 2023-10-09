FROM golang:1.21.2-alpine3.18

RUN mkdir /app
WORKDIR /app
COPY . .

# RUN go get
# RUN go build -o bin/gobank .
RUN make run 

EXPOSE 3000 50051

# ENTRYPOINT [ "./bin/gobank" ]