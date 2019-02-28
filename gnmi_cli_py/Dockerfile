FROM python:3

COPY . .
RUN pip install $(grep -ivE "futures" ./gnxi/gnmi_cli_py/requirements.txt)

WORKDIR "/gnxi/gnmi_cli_py"
RUN ["chmod", "+x", "/gnxi/gnmi_cli_py/py_gnmicli.py"]
CMD [ "python", "/gnxi/gnmi_cli_py/py_gnmicli.py" ]
