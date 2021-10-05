
app_name=$1

flogo create -f ${app_name} app
cd app
export GOOS=linux ;\
export GOARCH=amd64;\
flogo build -e --verbose;