'use strict';

var werks = angular.module('werks', ['ngResource', 'httpFix']);

werks.factory(
	'NewGame', function($resource) { return $resource('/api/newGame')});
werks.factory(
	'Game', function($resource) { return $resource('/api/game')});
werks.factory(
	'Locos', function($resource) { return $resource('/api/locos')});
werks.factory(
	'Players', function($resource) { return $resource('/api/players')});
werks.factory(
	'Message', function($resource) { return $resource('/api/message?g=:gameId')})

werks.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/board', {
		templateUrl: '/views/board.tmpl',
		controller: BoardCtrl
	});
	$routeProvider.when('/newGame', {
		templateUrl: '/views/newGame.tmpl',
		controller: NewGameCtrl
	});
}]);

var MainCtrl = function($scope, $route, $location, $routeParams, LocoSvc, GameSvc) {

	$scope.$route = $route;
  $scope.$location = $location;
  $scope.$routeParams = $routeParams;

  $scope.messages = [];

  $scope.getPlayers = function() {
  	return GameSvc.getPlayers();
  };

  $scope.generation = function(gen) {
  	return LocoSvc.getGeneration(gen);
  }
};

var NewGameCtrl = function($scope, $location, $http, NewGame, LocoSvc, GameSvc) {
	$scope.playerCount = 0;
	$scope.name = "New game";
	$scope.players = [
		{name: 'Player 1'},
		{name: 'Player 2'},
		{name: 'Player 3'},
		{name: 'Player 4'},
		{name: 'Player 5'}];

	$scope.startGame = function() {
		// build playerCount, player0, player1... form values
		var g = new NewGame();
		g.name = $scope.name;
		for (var i = 0; i < $scope.playerCount; i++) {
			var key = 'player' + i;
			g[key] = $scope.players[i].name;
		}
		g.playerCount = $scope.playerCount;

		// post to api/newGame, and on response, load the game data
		// and switch to the board view.
  	g.$save(function(data) {
  		console.log(data);
  		GameSvc.setGame(data);
  		$scope.$parent.locos = LocoSvc.buildLocosObject(data.locos);
  		$scope.$parent.rows = LocoSvc.buildRows(data.locos);
  		$scope.$parent.players = data.players;
			$location.path('board');
  	});
	}
}

var BoardCtrl = function($scope, $timeout, $http, Message, GameSvc) {

	$timeout(getMessage, 500);

	function getMessage() {
		var gameId = GameSvc.getGameId();
		Message.get({gameId: gameId}, function(data) {
			if (data) {
				$scope.messages.push(data);
			}
			$timeout(getMessage, 500);
		});
	}


}

var LocoCtrl = function($scope) {

	if ($scope.loco === undefined) {
		return;
	}

	var eo = $scope.loco.existingOrders;
	var cb = $scope.loco.customerBase;
	var io = $scope.loco.initialOrders;

	var rows = [
		[eo[4], null, null, null, null, null, cb[4]],
		[eo[3], null, null, null, null, null, cb[3]],
		[eo[2], null, null, null, null, null, cb[2]],
		[eo[1], eo[0], null, io, null, cb[0], cb[1]]
	];

	// fix the loco (T31) that has only one existing order
	// and customer base, moving its die spaces to the
	// corners.
	if ($scope.loco.maxExistingOrders == 1) {
		rows[3] = [eo[0], null, null, io, null, null, cb[0]];
	}

	$scope.rows = rows;

}

var FactoryCtrl = function($scope, FactorySvc) {

	var f = FactorySvc.getServiceInfo($scope);
	$scope.factoryServiceInfo = f;
	$scope.loco = f.loco;

	$scope.expandDisabled = function() {
		return FactorySvc.expand(true, $scope.factoryServiceInfo);
	}

	$scope.expand = function() {
		FactorySvc.expand(false, $scope.factoryServiceInfo);
	}

	$scope.upgradeDisabled = function() {
		return FactorySvc.upgrade(true, $scope.factoryServiceInfo);
	}

	$scope.upgrade = function() {
		FactorySvc.upgrade(false, $scope.factoryServiceInfo);
	}

	$scope.sellDisabled = function() {
		return FactorySvc.sell(true, $scope.factoryServiceInfo);
	};

	$scope.sell = function() {
		FactorySvc.sell(false, $scope.factoryServiceInfo);
	};
};

// put scrollToBottom on a div using ChatCtrl to have its content
// scroll to the bottom after a chat message gets added.
werks.directive('scrollToBottom', function($timeout) {
    return {
        link: function(scope, elm, attrs) {
            var box = elm[0];
            scope.$watch('chatMessages.length', function() {
                $timeout(function() {
                    box.scrollTop = box.scrollHeight;
                }, 250);
            }, false);
        }
    };
});

var ChatCtrl = function($scope, $http, $timeout, GameSvc) {
    $scope.text = 'chat';
    $scope.chatMessages = [];

		$scope.getMessage = function() {
			var gameId = GameSvc.getGameId();
			var playerId = GameSvc.getCurrentPlayerId();
			if (playerId == null) {
				return;
			}
			var url = '/api/chat?g=' + gameId + '&p=' + playerId;
			$http.get(url).success(function(data) {
				if (data) {
					console.log(data)
					$scope.chatMessages.push(data);
				}
				$timeout($scope.getMessage, 1000);
			});
		}

    $scope.sendChat = function() {
    	var gameId = GameSvc.getGameId();
    	var playerId = GameSvc.getCurrentPlayerId();
    	if (playerId == null) {
    		return;
    	}
    	var text = encodeURIComponent($scope.text);
			var url = '/api/chat?g=' + gameId + '&p=' + playerId + '&text=' + text;
      $http.post(url);
      $scope.text = ""
    }

		$timeout($scope.getMessage, 500);

};
