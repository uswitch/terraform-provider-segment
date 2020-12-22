version = 0.0.9
provider_path = uswitch.com/segment/segment/$(version)/darwin_amd64/

build: 
	go build -o terraform-provider-segment_$(version)

	mkdir -p ~/.terraform.d/plugins/$(provider_path)
	cp terraform-provider-segment_$(version) ~/.terraform.d/plugins/$(provider_path)
