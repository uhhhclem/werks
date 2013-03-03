'use strict';

var http;
var locoSvc;
var gameSvc;
var spy;

describe('GameSvc', function(){

	beforeEach(module('werks'));

	beforeEach(inject(function($injector) {
	  http = $injector.get('$httpBackend');
	  locoSvc = $injector.get('LocoSvc');
	  gameSvc = $injector.get('GameSvc');
	  spy = jasmine.createSpy('spy');
	  spy.callback = function() {};
	}));

	afterEach(function() {
		http.verifyNoOutstandingExpectation();
		http.verifyNoOutstandingRequest();
	});

	it('should return the globals dictionary.', function(){
		var g = gameSvc.getGlobals();
		expect(g.game).toBe(null);
		expect(g.username).toBe(null);
		expect(g.token).toBe(null);
	});

	it('should set user information.', function() {
		gameSvc.setUserInfo('username', 'token');
		var g = gameSvc.getGlobals();
		expect(g.username).toBe('username');
		expect(g.token).toBe('token');
	})

	it('should get the game.', function() {
		var g = gameSvc.getGlobals();
		g.token = 'token';

		var fake_game = {
			id: 'game_id',
			locos: 'locos',
			players: 'players'
		};
		spyOn(locoSvc, 'buildLocosObject').andCallFake(function(locos) { return '1';});
		spyOn(locoSvc, 'buildRows').andCallFake(function(locos) { return '2';})
		spyOn(spy, 'callback')

		http.expectGET('/api/game?g=game_id&u=token').respond(fake_game);
		gameSvc.getGame('game_id', spy.callback);
		http.flush();

		expect(g.game.id).toBe('game_id');
		expect(spy.callback).toHaveBeenCalled();
		expect(locoSvc.buildLocosObject).toHaveBeenCalledWith('locos');
		expect(locoSvc.buildRows).toHaveBeenCalledWith('locos');
		expect(g.locos).toBe('1');
		expect(g.rows).toBe('2');
	});

	it('should create a new game.', function() {
		var g = gameSvc.getGlobals();
		var fake_game = {
			id: 'game_id',
			locos: 'locos',
			players: 'players'
		};
		http.expectPOST(
				'/api/newGame',
				'playerCount=3&player0=test1&player1=test2&player2=test3')
				.respond(fake_game);

		spyOn(locoSvc, 'buildLocosObject').andCallFake(function(locos) { return '1';});
		spyOn(locoSvc, 'buildRows').andCallFake(function(locos) { return '2';})
		spyOn(spy, 'callback');

		gameSvc.newGame(['test1', 'test2', 'test3'], spy.callback);
		http.flush();

		expect(g.game.id).toBe('game_id');
		expect(spy.callback).toHaveBeenCalled();
		expect(locoSvc.buildLocosObject).toHaveBeenCalledWith('locos');
		expect(locoSvc.buildRows).toHaveBeenCalledWith('locos');
		expect(g.locos).toBe('1');
		expect(g.rows).toBe('2');
	});

	it('should get actions.', function(){
		var g = gameSvc.getGlobals();
		g.game = {
			id: 'game_id',
			players: [{id: 0, isCurrent: false}, {id: 1, isCurrent: true}]
		};
		http.expectGET('/api/action?g=game_id&p=1').respond('actions');
		gameSvc.getActions();
		http.flush();
		expect(g.actions).toBe('actions');
	});

	it('should post actions.', function() {
		var g = gameSvc.getGlobals();
		g.game = {
			id: 'game_id',
			players: [{id: 0, isCurrent: false}, {id: 1, isCurrent: true}]
		};
		http.expectPOST('/api/action?g=game_id&p=1&abbr=abbr')
			.respond({game: 'game', actions: 'actions'});
		gameSvc.doAction('abbr');
		http.flush();

		expect(g.game).toBe('game');
	});

});
