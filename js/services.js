'use strict';

werks.service('GameSvc', function() {

	this.setGame = function(game) {
		this._game = game;
	}

	this.getGameId = function() {
		return this._game.id;
	}

	this.getCurrentPlayer = function() {
		var game = this._game;
		for (var i=0; i<game.players.length; i++) {
			var p = game.players[i];
			if (p.isCurrent) {
				return p;
			}
		}
		return null;
	}

	this.getPlayers = function() {
		return this._game.players;
	}

	this.getCurrentPlayerId = function() {
		return this.getCurrentPlayer().id;
	}

});

werks.service('LocoSvc', function() {

	// build an object, keyed by loco key, out of the locos array
	this.buildLocosObject = function(locos) {
		var locoObject = {}
		for (var i = 0; i < locos.length; i++) {
			var loco = locos[i];
			locoObject[loco.key] = loco;
		}
		return locoObject;
	};

	// create an array that can be iterated over to generate the
	// table displaying the board
	this.buildRows = function(locos) {
		var indexes = [
			[0, 1, 2],
			[13, null, 3],
			[12, null, 4],
			[11, null, 5],
			[10, null, 6],
			[9, 8, 7]
		];

		var rows = []
		for (var i = 0; i < indexes.length; i++) {
			var row = indexes[i]
			rows.push([
					locos[row[0]],
					locos[row[1]],
					locos[row[2]],
				])
		}
		return rows;
	};

	// get the generation (Roman numeral) from the integer in
	// the data.
	this.getGeneration = function(gen) {
		return ['I', 'II', 'III', 'IV', 'V'][gen - 1];
	}
});

werks.service('FactorySvc', function () {

	this.getServiceInfo = function(scope) {
		var f = {};
		f.factory = scope.factory;
		f.player = scope.player;
		f.loco = scope.locos[scope.factory.key];
		f.upgradeTo = scope.locos[f.loco.upgradeTo];
		f.upgradeCost = (f.upgradeTo === undefined)
				? null
				: f.upgradeTo.productionCost - f.loco.productionCost;
		return f;
	}

	this.expand = function(checkOnly, f) {
		if (f.player.money < f.loco.productionCost) {
			return true;
		};
		if (checkOnly) {
			return false;
		}

		f.player.money -= f.loco.productionCost;
		f.factory.capacity += 1;
		f.factory.unitsSold = 0;
	};

	this.upgrade = function(checkOnly, f) {
		if (f.upgradeTo === undefined) {
			return true;
		}
		if (f.factory.capacity == 0) {
			return true;
		}
		var upgradeFactory = this._findFactory(f.player, f.upgradeTo)
		if (upgradeFactory == null) {
			return true;
		}
		if (f.upgradeCost > f.player.money) {
			return true;
		}

		if (checkOnly) {
			return false;
		}

		f.player.money -= f.upgradeCost;
		f.factory.capacity -= 1;
		upgradeFactory.capacity += 1;

		return false;
	};

	this.sell = function(checkOnly, f) {
		var unitsToSell = f.factory.capacity - f.factory.unitsSold;
		if (unitsToSell <= 0) {
			return true;
		}
		var die = this._findExistingOrder(f, unitsToSell);
		if (die == null || die.pips == 0) {
			return true;
		}
		if (checkOnly) {
			return false;
		}

		var unitsSold = 0;
		if (die.pips == unitsToSell) {
			unitsSold = unitsToSell;
		}
		else if (die.pips < unitsToSell) {
			unitsSold = die.pips;
		}
		else {
			unitsSold = unitsToSell;
		}

		if (die.pips > unitsSold) {
			die.pips -= unitsSold;
		}
		else {
			for (var i=0; i < f.loco.maxCustomerBase; i++) {
				var customerDie = f.loco.customerBase[i];
				if (customerDie.pips == 0) {
					customerDie.pips = die.pips;
					die.pips = 0;
					break;
				}
			}
		}

		f.player.money += unitsSold * f.loco.income;
		f.factory.unitsSold += unitsSold;
		f.factory.capacity -= unitsSold;

		return false;
	}

	this._findFactory = function(player, loco) {
		for (var i = 0; i < player.factories.length; i++)
		{
			var factory = player.factories[i];
			if (factory.key == loco.key) {
				return factory;
			}
		}
		return null;
	};

	// find the first existing-order die that's exactly big enough
	// to fulfill the order, or else the highest die.
	this._findExistingOrder = function(f, unitsToSell) {
		var die = null;
		for (var i = 0; i < f.loco.maxExistingOrders; i++) {
			var order = f.loco.existingOrders[i];
			if (order.pips == unitsToSell) {
				return order;
			}
			if (die == null || die.pips < order.pips) {
				die = order;
			}
		}
		return die;
	}

});
