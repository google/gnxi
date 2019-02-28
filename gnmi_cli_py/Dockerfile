FROM python:3

# This is assuming you're within the git cloned gnxi repo.
COPY . .
RUN pip install $(grep -ivE "futures" ./requirements.txt)

WORKDIR "/gnxi/gnmi_cli_py"
RUN ["chmod", "+x", "/gnxi/gnmi_cli_py/py_gnmicli.py"]
CMD [ "python", "/gnxi/gnmi_cli_py/py_gnmicli.py" ]
