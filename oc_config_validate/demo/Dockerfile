FROM python:3.9-slim-bullseye

RUN apt-get update
RUN apt-get install -y procps
RUN pip3 install --extra-index-url https://test.pypi.org/simple/ oc-config-validate

WORKDIR /opt/oc_config_validate

COPY ./* demo/

CMD /opt/oc_config_validate/demo/run_demo.sh -L -N
