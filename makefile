.PHONY: gen
gen:
	go get -u ./...
	go mod tidy
	go generate ./...
	go test -v ./...

.PHONY: gen-helm
gen-helm:
	helm dependency update ./charts/icarus
	helm dependency build ./charts/icarus

.PHONY: helm
helm: gen-helm
	helm install icarus ./charts/icarus

.PHONY: uhelm
uhelm: gen-helm
	helm upgrade icarus ./charts/icarus

.PHONY: secret
secret:
	kubectl create secret generic tunnel-credentials --from-file=credentials.json=${HOME}/.cloudflared/${CLOUDFLARE_TUNNEL_ID}.json


REGISTRY=njayp
IMAGE=${REGISTRY}/hermes

.PHONY: image
image: 
	docker build -t ${IMAGE} .

.PHONY: run
run: image
	docker run -it --rm --name cf \
	-v ${HOME}/.cloudflared/cert.pem:/home/nonroot/cert.pem \
	-v ${HOME}/repos/hermes/ingress.yml:/home/nonroot/ingress.yml \
	-e CLOUDFLARE_API_KEY=${CLOUDFLARE_API_KEY} \
	-e CLOUDFLARE_EMAIL=${CLOUDFLARE_EMAIL} \
	${IMAGE}

.PHONY: push
push: image
	docker push ${IMAGE}

