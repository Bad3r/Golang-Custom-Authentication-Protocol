FROM python:3.9.0

# Same thing than the 'server' Dockerfile.
# We import 'client.py' in '/client/' folder.
COPY client.py /client/

# I would like to introduce something new, the 'WORKDIR' command.
# This command changes the base directory of your image.
# Here we define '/client/' as base directory.
WORKDIR /client/
