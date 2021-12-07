
app_name=$1

flogo create -f ${app_name} app
cd app
flogo build -e --verbose;