'use strict';

var werks = angular.module('werks', ['ngResource', 'httpFix']);

werks.factory(
	'Action', function($resource) {return $resource(
				'/api/action')});
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
	$routeProvider.when('/login', {
		templateUrl: '/views/login.tmpl',
		controller: LoginCtrl
	});
	$routeProvider.when('/board', {
		templateUrl: '/views/board.tmpl',
		controller: BoardCtrl
	});
	$routeProvider.when('/newGame', {
		templateUrl: '/views/newGame.tmpl',
		controller: NewGameCtrl
	});
	$routeProvider.otherwise({ redirectTo: 'login'});
}]);


var MainCtrl = function($scope, $route, $location, $routeParams, LocoSvc, GameSvc, Action) {

	$scope.$route = $route;
  $scope.$location = $location;
  $scope.$routeParams = $routeParams;

  $scope.globals = GameSvc.getGlobals();

  $scope.messages = [];

  $scope.generation = function(gen) {
  	return LocoSvc.getGeneration(gen);
  }

  $scope.getActions = function() {
  	var a = new Action();
  	a.$get(GameSvc.urlParams(), function(data) {
  		$scope.currentPlayer = GameSvc.getCurrentPlayer();
  		$scope.actions = data;
  	})
  }
};

var LoginCtrl = function($scope, $location, $http, GameSvc) {

	$scope.registering = false;
	$scope.errorMessage = null;

	$scope.toggleRegistering = function() {
		$scope.registering = !$scope.registering;
	}

	$scope.registerDisabled = function() {
		var p1 = $scope.password;
		var p2 = $scope.repeatPassword;

		return !p1 || !p2 || (p1 != p2) || p1.length < 4;
	}

	$scope.login = function() {
		$scope.errorMessage = null;
		var url = '/api/login?u=' + $scope.username + '&p=' + $scope.password;
		$http.get(url).success(function(data){
			console.log(data);
			if (data.token) {
				GameSvc.setUserInfo($scope.username, data.token);
				$location.path('/newGame');
			} else {
				$scope.errorMessage = 'Invalid login, try again.';
			}
		})
	}

	$scope.register = function() {
		if ($scope.registerDisabled()) {
			return;
		}
		$scope.errorMessage = null;
		var url = '/api/register?u=' + $scope.username + '&p=' + $scope.password;
		$http.get(url).success(function(data){
			console.log(data);
			if (data.token) {
				$scope.user.username = $scope.username
				$scope.user.token = data.token;
				$location.path('/newGame');
			} else {
				$scope.errorMessage = data.msg;
			}
		})
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
		var players = [];
		for (var i = 0; i < $scope.playerCount; i++) {
			players.push($scope.players[i].name);
			GameSvc.newGame(players, function() {
				$location.path('board');
			});
		}
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

var ActionCtrl = function($scope, GameSvc, Action) {

	$scope.player = GameSvc.getCurrentPlayer();

	var arr = $scope.a.abbr.split(":");
	$scope.actionType = arr[0];
	if (arr.length > 1) {
		$scope.locoKey = arr[1];
		var loco = GameSvc.getLoco($scope.locoKey);
		$scope.locoKind = loco.kind;
	} else {
		$scope.locoKey = null;
		$scope.locoType = null;
	}

	$scope.doAction = function() {
		var a = new Action();
		a.abbr = $scope.a.abbr;
		a.$save(GameSvc.urlParams(), function(data) {
			console.log(data);
		})
	}

};
