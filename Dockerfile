FROM golang:1.21-alpine as builder

COPY ./src /src/
WORKDIR /src

RUN go build -o PodResourceCalculator

FROM scratch
COPY --from=builder /src/PodResourceCalculator /

#CMD [ "sleep", "infinity" ]
