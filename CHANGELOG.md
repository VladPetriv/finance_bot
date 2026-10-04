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
