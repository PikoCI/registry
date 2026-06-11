// router.js - Backbone Router
(function() {
  'use strict';

  PkReg.Router = Backbone.Router.extend({
    routes: {
      '':                          'home',
      'search(?*qs)':              'search',
      'plugins/:ns/:name':         'typeDetail',
      'plugins/:ns':               'namespace',
      'tags':                      'tags',
      'tags/:tag':                 'tagDetail',
      'login':                     'login',
      'login/callback(?*qs)':       'loginCallback',
      'me/types':                  'myTypes',
      'me/tokens':                 'tokens',
      '*notFound':                 'home'
    },

    initialize: function(options) {
      this.mainView = options.mainView;
    },

    // Parse query string helper
    _parseQS: function(qs) {
      var params = {};
      if (!qs) return params;
      qs = qs.replace(/^\?/, '');
      _.each(qs.split('&'), function(pair) {
        var parts = pair.split('=');
        if (parts.length === 2) {
          params[decodeURIComponent(parts[0])] = decodeURIComponent(parts[1]);
        }
      });
      return params;
    },

    home: function() {
      this.mainView.showContent(new PkReg.Views.SearchView({
        query: '',
        kind: '',
        page: 1
      }));
    },

    search: function(qs) {
      var params = this._parseQS(qs);
      this.mainView.showContent(new PkReg.Views.SearchView({
        query: params.q || '',
        kind: params.kind || '',
        page: parseInt(params.page, 10) || 1
      }));
    },

    typeDetail: function(ns, name) {
      this.mainView.showContent(new PkReg.Views.TypeDetailView({
        namespace: ns,
        name: name
      }));
    },

    namespace: function(ns) {
      this.mainView.showContent(new PkReg.Views.NamespaceView({
        namespace: ns
      }));
    },

    tags: function() {
      this.mainView.showContent(new PkReg.Views.TagsView());
    },

    tagDetail: function(tag) {
      this.mainView.showContent(new PkReg.Views.TagDetailView({
        tagName: tag
      }));
    },

    login: function() {
      if (PkReg.Auth.isLoggedIn()) {
        Backbone.history.navigate('me/types', { trigger: true });
        return;
      }
      this.mainView.showContent(new PkReg.Views.LoginView());
    },

    loginCallback: function(qs) {
      var params = this._parseQS(qs);
      var code = params.code;
      if (!code) {
        Backbone.history.navigate('login', { trigger: true });
        return;
      }
      var self = this;
      $.ajax({
        url: '/api/auth/github',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ code: code }),
        dataType: 'json',
        success: function(resp) {
          PkReg.Auth.setToken(resp.token);
          PkReg.Auth.setUser(resp.user);
          if (self.mainView) {
            self.mainView.refreshHeader();
          }
          Backbone.history.navigate('me/types', { trigger: true });
        },
        error: function(xhr) {
          var msg = 'Login failed';
          try { msg = JSON.parse(xhr.responseText).error; } catch(e) {}
          alert(msg);
          Backbone.history.navigate('login', { trigger: true });
        }
      });
    },

    myTypes: function() {
      this.mainView.showContent(new PkReg.Views.MyTypesView());
    },

    tokens: function() {
      this.mainView.showContent(new PkReg.Views.TokensView());
    }
  });

})();
