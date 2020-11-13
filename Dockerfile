FROM ubuntu:latest
WORKDIR /code
COPY kds .
ENTRYPOINT [ "/code/kds", \
    "-v", "2", \
    "-logtostderr", \
    "--username", "dev", \
    "--password", "dev", \
    "--host", "db", \
    "--nodeUri", "http://121.89.211.107:34568", \
    "--httpPort", "8083" \
     ]