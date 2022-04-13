# [0.4.0](https://github.com/ThoronicLLC/collector/compare/v0.3.1...v0.4.0) (2022-04-13)


### Features

* **processor:** added key-value and CEF processor ([#49](https://github.com/ThoronicLLC/collector/issues/49)) ([9bbce68](https://github.com/ThoronicLLC/collector/commit/9bbce683468a201203e30c73c51e69b0e876ddaa))
* **processor:** added syslog message processor ([#48](https://github.com/ThoronicLLC/collector/issues/48)) ([807e2cc](https://github.com/ThoronicLLC/collector/commit/807e2cc2355c5637b1228d6853f9ba7c4416276f))



## [0.3.1](https://github.com/ThoronicLLC/collector/compare/v0.3.0...v0.3.1) (2022-03-25)


### Bug Fixes

* correctly flush sqs input results when closed ([#45](https://github.com/ThoronicLLC/collector/issues/45)) ([7184a73](https://github.com/ThoronicLLC/collector/commit/7184a739a63cae55b0a890496658e483e48d20b7))
* **input:** correctly flush data when pubsub input is closed ([#42](https://github.com/ThoronicLLC/collector/issues/42)) ([6374665](https://github.com/ThoronicLLC/collector/commit/637466579a04876803e183dcb163ca1c57b2f1d8))



# [0.3.0](https://github.com/ThoronicLLC/collector/compare/v0.2.0...v0.3.0) (2022-03-23)


### Bug Fixes

* prevent junk directories existing in release zip ([#37](https://github.com/ThoronicLLC/collector/issues/37)) ([6b0d835](https://github.com/ThoronicLLC/collector/commit/6b0d8350717ce916abcc198ee6fe8fca1d975727))


### Features

* added file output plugin ([#39](https://github.com/ThoronicLLC/collector/issues/39)) ([37c8b78](https://github.com/ThoronicLLC/collector/commit/37c8b7878d5d70dac3b2c070664954a21415d62b))



# [0.2.0](https://github.com/ThoronicLLC/collector/compare/688fc2b9d86398b715ef50d49c75040e1c52da05...v0.2.0) (2022-03-23)


### Bug Fixes

* overhauled error, status, and tmp file management ([5dee582](https://github.com/ThoronicLLC/collector/commit/5dee582f6d28f2d8bd2a9e14b06d82c479306750))
* prevent empty file from being sent to processors ([53f35b4](https://github.com/ThoronicLLC/collector/commit/53f35b44a0215c4d5e22352c1237c96bae69e08a))
* updated manager to correctly initialize processors ([5795a51](https://github.com/ThoronicLLC/collector/commit/5795a513a8c4fa9e6ff839151ad64d5b63e06535))


### Features

* added AWS SQS input ([#31](https://github.com/ThoronicLLC/collector/issues/31)) ([6762536](https://github.com/ThoronicLLC/collector/commit/6762536ea1ad74d5717b36a47b77a4e258054baa))
* added cel processor ([8ba0d74](https://github.com/ThoronicLLC/collector/commit/8ba0d745bb27f130b0b766e9d0167e1e20aae613))
* added Microsoft Graph security alerts input ([#35](https://github.com/ThoronicLLC/collector/issues/35)) ([f83c1b2](https://github.com/ThoronicLLC/collector/commit/f83c1b2429e48596f6144f73545fe105b8601df6))
* added plugin validation and tests ([#28](https://github.com/ThoronicLLC/collector/issues/28)) ([8e40b68](https://github.com/ThoronicLLC/collector/commit/8e40b68b6e2f1947daaf3b8a95f768f7947b1dbb))
* **cli:** overhauled the CLI application to be useful ([#4](https://github.com/ThoronicLLC/collector/issues/4)) ([b81b2bd](https://github.com/ThoronicLLC/collector/commit/b81b2bd51618ccdfe298a75fae6af2e82d3d555a))
* **core:** added variable replacer helper function for plugins ([#12](https://github.com/ThoronicLLC/collector/issues/12)) ([0bb8f59](https://github.com/ThoronicLLC/collector/commit/0bb8f5943f13f119f117b50b4d59f61bf8ade946))
* create github issue templates ([d96d4c1](https://github.com/ThoronicLLC/collector/commit/d96d4c10018b73348f6a56a86ec347384b87388d))
* initial commit ([688fc2b](https://github.com/ThoronicLLC/collector/commit/688fc2b9d86398b715ef50d49c75040e1c52da05))
* **input:** added Google Cloud PubSub input ([#16](https://github.com/ThoronicLLC/collector/issues/16)) ([68ba0b9](https://github.com/ThoronicLLC/collector/commit/68ba0b9a079894e4591e776c390c07de3e2e24e1))
* **input:** added syslog input ([#17](https://github.com/ThoronicLLC/collector/issues/17)) ([4bdc07f](https://github.com/ThoronicLLC/collector/commit/4bdc07fe3181807263371a7ffd68b74ee0e53a46))
* **input:** support file deletion in file input plugin ([23c69c2](https://github.com/ThoronicLLC/collector/commit/23c69c200e8423b9f173ec48c609fc6eb0db9197))
* **output:** added Google Cloud PubSub output ([#24](https://github.com/ThoronicLLC/collector/issues/24)) ([0c12bd9](https://github.com/ThoronicLLC/collector/commit/0c12bd950a7806d8e9f5e35c8f67e00922d73952))
* **output:** added google cloud storage output ([#20](https://github.com/ThoronicLLC/collector/issues/20)) ([2dd0c1c](https://github.com/ThoronicLLC/collector/commit/2dd0c1cd9b9c583ec6554c0717839bf757c655e8))
* **output:** added Microsoft Azure Log Analytics as an output ([#22](https://github.com/ThoronicLLC/collector/issues/22)) ([7c5c37a](https://github.com/ThoronicLLC/collector/commit/7c5c37a8af84e1798233edc6c3febe73d452d0e9))
* **output:** Amazon S3 Output ([#2](https://github.com/ThoronicLLC/collector/issues/2)) ([ceca476](https://github.com/ThoronicLLC/collector/commit/ceca476e5954698e1642e192e26aa79f7df7183f))



