# zeitgeber
A framework that provides an apple-to-apple comparison for state-of-the-art chained BFT SMR.
HotStuff introduced chained framework into BFT SMR and classified into two categories, two-chain protocols (PBFT, Tendermint, Casper), three-chain protocols
(HotStuff, LibraBFT), according to their commit rule.
These protocols have different safety rules (voting, commit) and liveness rules and therefore, should have varying performance under different conditions, especially under performance-failure attacks.
This gives us a chance to build a general framework to easily implement these protocols using the same primitives, only leaving safety rules and liveness rules for developers.
