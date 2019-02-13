FROM scratch
EXPOSE 8080
EXPOSE 8090
ENTRYPOINT ["/locations"]
COPY ./bin/ /
