# [0.5.0](https://github.com/ThoronicLLC/collector/compare/v0.4.1...v0.5.0) (2022-12-21)


### Features

* added kafka input and output plugins ([#54](https://github.com/ThoronicLLC/collector/issues/54)) ([e14403f](https://github.com/ThoronicLLC/collector/commit/e14403f2bd3b6592af9193961bb125a22443605e))



## [0.4.1](https://github.com/ThoronicLLC/collector/compare/v0.4.0...v0.4.1) (2022-09-08)


### Bug Fixes

* refactor tmp_writer to not open files until data is being written ([#51](https://github.com/ThoronicLLC/collector/issues/51)) ([1228c79](https://github.com/ThoronicLLC/collector/commit/1228c798f57ac211da7ec733b02e83b7e8b73c80))



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



