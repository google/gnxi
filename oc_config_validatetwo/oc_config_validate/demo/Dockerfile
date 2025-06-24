FROM python:3.9-slim-bullseye

RUN apt-get update
RUN apt-get install -y procps
RUN pip3 install oc-config-validate

WORKDIR /opt/oc_config_validatetwo

COPY ./* demo/

CMD /opt/oc_config_validatetwo/demo/run_demo.sh -L -N
