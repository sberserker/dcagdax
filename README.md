# DCA Coinbase

Automated dollar cost averaging for BTC, LTC, BCH and ETH on GDAX.
Inspired by https://github.com/blampe/dcagdax
Added the following features
- automatic deposits
- fixed ledger api dependency package. Had to move the package inside as well
- added percentage buys
- added force flag to buy now regardless of the window check, will ask for confirmation

## Setup

If you only have a Coinbase account you'll need to also sign into
[Coinbase](https://pro.coinbase.com/). Make sure you have a bank account linked to one of these for
ACH transfers.

Procure a Coinbase API key for yourself by visiting
[https://pro.coinbase.com/profile/api](https://pro.coinbase.com/profile/api). **Do not share
this API key with third parties!**

## Usage

Build the binary:

```
$ go build -o .  ./...
```

Then run it:

```
./dcagdax --help
usage: dcagdax --every=EVERY [<flags>]

Flags:
  --help         Show context-sensitive help (also try --help-long and
                 --help-man).
  --coin=BTC     Which coin you want to buy: BTC, LTC, BCH or ETH : percentage amount. Can be split between multipe coins. Total must be 100%. Example --coin BTC:70 --coin ETH:30
  --every=EVERY  How often to make purchases, e.g. 1h, 7d, 3w.
  --usd=USD      How much USD to spend on each purchase. If unspecified, the
                 minimum purchase amount allowed will be used.
  --until=UNTIL  Stop executing trades after this date, e.g. 2017-12-31.
  --trade        Actually execute trades.
  --autofund     Automatically initiate ACH deposits.
  --force        Force trade despite trading windows, will ask for user confirmation
  --version      Show application version.
```

Run the `dcagdax` binary with an environment containing your API credentials:
```
$ GDAX_SECRET=secret \
  GDAX_KEY=key \
  GDAX_PASSPHRASE=pass \
  ./dcagdax --help
```

Be aware that if you set your purchase amount near 0.01 BTC (the minimum trade
amount) then an upswing in price might prevent you from trading.

## FAQ

**Q:** Why do I not see any trading activity from the bot?

**A:** If you have other BTC trades on your account, the bot will detect that as a
cost-averaged purchase and wait until the next purchase window. This is for
people who want to "set it and forget it," not day traders!

**Q:** Why would I use this instead of Coinbase's recurring purchase feature?

**A:** The [fees on recurring
purchases](https://support.coinbase.com/customer/portal/articles/2109597)
(currently a minimum of $2.99 per purchase!) can add up quickly. This
side-steps those costs by automating free ACH deposits into your exchange
account & submitting market orders to exchange with BTC.

**Q:** How should I deploy this?

**A:** You could run this as a periodic cronjob on your workstation or in the
cloud. Just be sure your API key & secret are not made available to anyone else
as part of your deployment!

**Q:** Can this auto-withdraw coins into a cold wallet?

**A:** Yes

**Q:** Which coins can I purchase?

**A:** We support all of Coinbase's products: BTC, LTC, BCH and ETH.
