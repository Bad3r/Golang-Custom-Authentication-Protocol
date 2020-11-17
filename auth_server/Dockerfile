FROM python:3.9.0

# In order to launch our python code, we must import the 'server.py' and 'index.html' file.
# We use the keyword 'ADD' to do that.
# Just a remember, the first parameter 'server.py' is the name of the file on the host.
# The second parameter '/auth_server/' is the path where to put the file on the image.
# Here we put files at the image '/auth_server/' folder.
COPY server.py /auth_server/
COPY index.html /auth_server/

# I would like to introduce something new, the 'WORKDIR' command.
# This command changes the base directory of your image.
# Here we define '/auth_server/' as base directory (where all commands will be executed).
WORKDIR /auth_server/