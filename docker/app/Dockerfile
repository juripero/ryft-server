FROM ubuntu:16.10

RUN apt-get update -y && apt-get install -y \
     dpkg \
     tar \
     less \
     vim \
     jq \
     curl

RUN groupadd ryftuser && \
    useradd -g ryftuser ryftuser

EXPOSE 8765

COPY .tmp/ryft-server /ryft-server
COPY etc/ryft-server.conf /etc/ryft-server.conf
COPY etc/ryft-users.yaml /etc/ryft-users.yml

RUN chmod +x /ryft-server

CMD /ryft-server --config=/etc/ryft-server.conf --address :8765 --debug