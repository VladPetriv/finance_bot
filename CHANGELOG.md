# [0.11.0](https://github.com/VladPetriv/finance_bot/compare/v0.10.0...v0.11.0) (2026-10-04)


### Features

* add docker-compose + makefile commands + dot env example ([4013568](https://github.com/VladPetriv/finance_bot/commit/4013568c4fcb075cc69bdacc941b120b407ab09b))
* automatic report engine ([81862eb](https://github.com/VladPetriv/finance_bot/commit/81862eb18f9f515786245af55605456fe8303e20))
* bump up golang version to 1.27.1 ([025d83d](https://github.com/VladPetriv/finance_bot/commit/025d83d05643404ec38a243b69a42d315dab6c84))
* **keyboard:** add base keyboard templates for automatic report ([ed32b96](https://github.com/VladPetriv/finance_bot/commit/ed32b96a6542c6d1773de49750b92a83ed7ce0b8))
* **migration/automatic_reports:** add more enum values for report period ([c49d9f5](https://github.com/VladPetriv/finance_bot/commit/c49d9f588cdf72909a5b24f476dfdaf8da2a7b0d))
* **migration:** init automatic reports ([30ff03a](https://github.com/VladPetriv/finance_bot/commit/30ff03aa894fe3b9c2859efe76ef73f2229b2bc3))
* **migrations:** add timezone for user, migrate all time related columns to timestampz ([47638c8](https://github.com/VladPetriv/finance_bot/commit/47638c872f8cd385b7d04569a10af95c27aaeaea))
* **migrator:** add ability to enable/disable logger ([9b15e14](https://github.com/VladPetriv/finance_bot/commit/9b15e1492bb552cc28764ab30c39c38dea2e3796))
* **model/state:** add automatic report events to state.GetEvent ([429bec0](https://github.com/VladPetriv/finance_bot/commit/429bec091bea06799e4ba4317d3211c06bc879d8))
* **model:** add base command/event/flow/flow steps for automatic ([f729a12](https://github.com/VladPetriv/finance_bot/commit/f729a129a3c71266f9a20e94443896775a2f0617))
* **model:** add new commands/flow/steps for timezone update ([4913f41](https://github.com/VladPetriv/finance_bot/commit/4913f41a4fdd2a867e3055c399bbcd6fad4cd178))
* **model:** adjust everything that needed for new commands and flows ([ceb1cf0](https://github.com/VladPetriv/finance_bot/commit/ceb1cf083935faff211e9d81f114ee1fe720f349))
* **model:** adopt layer to work with timezone ([08c84c2](https://github.com/VladPetriv/finance_bot/commit/08c84c286446ad36b8d6c7dfe2976c0b4468ddb3))
* **model:** init models for automatic reports logic ([fc41328](https://github.com/VladPetriv/finance_bot/commit/fc413282bd9902d7a9f3fcc9cd177dc9c08da5c8))
* **model:** init timezone model ([27ed229](https://github.com/VladPetriv/finance_bot/commit/27ed229462b1211726559b5926d6bb815d4bfafa))
* **pkg/database:** add utc timezone for postgres conn ([e1926bf](https://github.com/VladPetriv/finance_bot/commit/e1926bf02e53e1d50f481cf3dd4d287ee4315a78))
* **pkg/logger:** add dummy logger ([1a9ed35](https://github.com/VladPetriv/finance_bot/commit/1a9ed35788fce1dd53269f57f29eb1c98b1720e8))
* **service:** add keyboards for automatic reports, add selection keyboard ([e44f404](https://github.com/VladPetriv/finance_bot/commit/e44f404db1ff5f2d29a381200cc1fe2c0e09e1e3))
* **service:** add keyboards for work with timezones ([f749aeb](https://github.com/VladPetriv/finance_bot/commit/f749aeb49c3c729e5daa690e7f707e3db9edf8ec))
* **service:** add select all/unselect all button for keyboard ([adbff9a](https://github.com/VladPetriv/finance_bot/commit/adbff9a8f702636ba598b5f6eab36cbfcce70740))
* **service:** automatic reports crud ([612c732](https://github.com/VladPetriv/finance_bot/commit/612c732021d9aa518300506465c18c46daf54221))
* **service:** integrate timezone ([7c0af79](https://github.com/VladPetriv/finance_bot/commit/7c0af79e29a434cb7e131aa7707e92354fd1bad0))
* **service:** new keyboard template for automatic report ([556f9f9](https://github.com/VladPetriv/finance_bot/commit/556f9f95d17b30a07fbb29e9aa222fedd361284c))
* **state:** add new automatic report event to simple events ([326a6ea](https://github.com/VladPetriv/finance_bot/commit/326a6ea76d0f6d8bc7b6e206849f58aa5db96048))
* **store/operation:** add between filter ([ebbc252](https://github.com/VladPetriv/finance_bot/commit/ebbc2529f068d7be114a8b34029765308f4ece39))
* **store:** add automatic report store ([3082904](https://github.com/VladPetriv/finance_bot/commit/30829040aeae6b9f897cd1c9c5ba82bbbbbc7595))
* **store:** adopt storage for timezone ([f642aeb](https://github.com/VladPetriv/finance_bot/commit/f642aeb5b08eb9b5119558e1c43886fe83b81ae7))

# [0.10.0](https://github.com/VladPetriv/finance_bot/compare/v0.9.10...v0.10.0) (2026-10-04)


### Features

* add docker-compose + makefile commands + dot env example ([6bed235](https://github.com/VladPetriv/finance_bot/commit/6bed23510337e5b4f148e1398c53935f96b3bc3f))
* automatic report engine ([fd1fc08](https://github.com/VladPetriv/finance_bot/commit/fd1fc083629b1b5b1c571973712a5f9dd7163408))
* bump up golang version to 1.27.1 ([34657db](https://github.com/VladPetriv/finance_bot/commit/34657db1de536b7689cc4dba309921179771f7f8))
* **keyboard:** add base keyboard templates for automatic report ([439c40b](https://github.com/VladPetriv/finance_bot/commit/439c40b76f29bceeec5ee0d44b3988962d5e902b))
* **migration/automatic_reports:** add more enum values for report period ([072bdda](https://github.com/VladPetriv/finance_bot/commit/072bdda548380ef8f9a77c3bb9eb94adaf998fe1))
* **migration:** init automatic reports ([f7bd6e1](https://github.com/VladPetriv/finance_bot/commit/f7bd6e18d98916aeecec85271f4190e4b529fc9b))
* **migrations:** add timezone for user, migrate all time related columns to timestampz ([1437869](https://github.com/VladPetriv/finance_bot/commit/1437869b68a49967971e8676b8a28a49868a6c12))
* **migrator:** add ability to enable/disable logger ([b9f403e](https://github.com/VladPetriv/finance_bot/commit/b9f403eff5f21c968bd40a7138e9d423207e53a6))
* **model/state:** add automatic report events to state.GetEvent ([3d70bbc](https://github.com/VladPetriv/finance_bot/commit/3d70bbc2a7d471364e3131c7a3e5d8a5c612fe3e))
* **model:** add base command/event/flow/flow steps for automatic ([c453a3f](https://github.com/VladPetriv/finance_bot/commit/c453a3ff368ce3b628b45c9589eb00c52d1c307f))
* **model:** add new commands/flow/steps for timezone update ([b11dd16](https://github.com/VladPetriv/finance_bot/commit/b11dd160a718b48aa178ef667771e13466969762))
* **model:** adjust everything that needed for new commands and flows ([fa648cc](https://github.com/VladPetriv/finance_bot/commit/fa648ccd2b1e6d96e712f55448a35edda631d4e7))
* **model:** adopt layer to work with timezone ([6065f92](https://github.com/VladPetriv/finance_bot/commit/6065f929c0b752cb035e983ef5c53cd42ff58cd6))
* **model:** init models for automatic reports logic ([287c7ab](https://github.com/VladPetriv/finance_bot/commit/287c7ab445700c8c1ed2837d35953621c7fba088))
* **model:** init timezone model ([265d9f9](https://github.com/VladPetriv/finance_bot/commit/265d9f98190d99e26ea14dd67b17eb88cc966891))
* **pkg/database:** add utc timezone for postgres conn ([aec0884](https://github.com/VladPetriv/finance_bot/commit/aec08843682426565f03c8ada25a902f20d9cc28))
* **pkg/logger:** add dummy logger ([2b33041](https://github.com/VladPetriv/finance_bot/commit/2b33041e348305b4614baafaef5fde6a92e498e3))
* **service:** add keyboards for automatic reports, add selection keyboard ([5af8770](https://github.com/VladPetriv/finance_bot/commit/5af877020a43f1124aa5b6ba474c697c776a84c5))
* **service:** add keyboards for work with timezones ([e7f0c0d](https://github.com/VladPetriv/finance_bot/commit/e7f0c0db31a7f9070da8fe69957748707ad98fb2))
* **service:** automatic reports crud ([ad18237](https://github.com/VladPetriv/finance_bot/commit/ad18237bce1bca3a4dd1cbf76ac704d7f6803d47))
* **service:** integrate timezone ([edc05f1](https://github.com/VladPetriv/finance_bot/commit/edc05f144ce3493457d23b8a2e336c90c8498cfb))
* **service:** new keyboard template for automatic report ([7bc6386](https://github.com/VladPetriv/finance_bot/commit/7bc638616758aed5a12f94537f1ca841c9973c31))
* **state:** add new automatic report event to simple events ([c4fe1ed](https://github.com/VladPetriv/finance_bot/commit/c4fe1edd6f763d9d101ab01f76f673bddd716ee1))
* **store/operation:** add between filter ([97d7fa6](https://github.com/VladPetriv/finance_bot/commit/97d7fa6d7009ef3996f507af867e369cff119cee))
* **store:** add automatic report store ([e024b29](https://github.com/VladPetriv/finance_bot/commit/e024b299001885c017cfbb985eca75066c942096))
* **store:** adopt storage for timezone ([f93fed0](https://github.com/VladPetriv/finance_bot/commit/f93fed06fa18da2ed5ea2abb94851bedb1c9dcc0))

## [0.9.10](https://github.com/VladPetriv/finance_bot/compare/v0.9.9...v0.9.10) (2026-08-02)


### Bug Fixes

* **telegram:** do not use emptyMessage trick, send Loading... message instead ([71b0f82](https://github.com/VladPetriv/finance_bot/commit/71b0f82e131794e821403591fa044cdcfee15c5c))
