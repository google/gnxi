FROM python:3.9-slim-bullseye

RUN apt-get update
RUN apt-get install -y git procps

COPY --from=golang:1.18-bullseye /usr/local/go/ /usr/local/go/
ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /opt
RUN git clone https://github.com/google/gnxi.git
RUN pip3 install -r ./gnxi/oc_config_validate/requirements.txt

WORKDIR /opt/gnxi/oc_config_validate
COPY run_demo.sh demo/run_demo.sh

CMD /opt/gnxi/oc_config_validate/demo/run_demo.sh -L
