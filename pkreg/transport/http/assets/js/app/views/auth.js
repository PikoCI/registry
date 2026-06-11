// views/auth.js - Auth views: login, token management, my types
(function() {
  'use strict';

  PkReg.Views = PkReg.Views || {};

  // Login page
  PkReg.Views.LoginView = Backbone.View.extend({
    template: _.template($('#tmpl-login').html()),

    events: {
      'click .btn-github': 'onGitHubLogin'
    },

    render: function() {
      this.$el.html(this.template({}));
      return this;
    },

    onGitHubLogin: function(e) {
      e.preventDefault();
      window.location.href = '/api/auth/github';
    }
  });

  // My Types page
  PkReg.Views.MyTypesView = Backbone.View.extend({
    template: _.template($('#tmpl-my-types').html()),

    initialize: function() {
      if (!PkReg.Auth.isLoggedIn()) {
        Backbone.history.navigate('login', { trigger: true });
        return;
      }
      this.types = [];
      this.fetchTypes();
    },

    render: function() {
      this.$el.html(this.template({
        types: this.types,
        formatKind: PkReg.formatKind,
        kindIcon: PkReg.kindIcon,
        formatDate: PkReg.formatDate,
        user: PkReg.Auth.getUser()
      }));
      return this;
    },

    fetchTypes: function() {
      var self = this;
      $.ajax({
        url: '/api/me/types',
        method: 'GET',
        dataType: 'json',
        headers: PkReg.Auth.ajaxHeaders(),
        success: function(resp) {
          self.types = resp || [];
          self.render();
        },
        error: function(xhr) {
          if (xhr.status === 401) {
            PkReg.Auth.logout();
            Backbone.history.navigate('login', { trigger: true });
          }
        }
      });
    }
  });

  // Token management page
  PkReg.Views.TokensView = Backbone.View.extend({
    template: _.template($('#tmpl-tokens').html()),

    events: {
      'submit .create-token-form': 'onCreate',
      'click .revoke-token-btn': 'onRevoke'
    },

    initialize: function() {
      if (!PkReg.Auth.isLoggedIn()) {
        Backbone.history.navigate('login', { trigger: true });
        return;
      }
      this.tokens = [];
      this.newToken = null;
      this.fetchTokens();
    },

    render: function() {
      this.$el.html(this.template({
        tokens: this.tokens,
        newToken: this.newToken,
        formatDate: PkReg.formatDate
      }));
      return this;
    },

    fetchTokens: function() {
      var self = this;
      $.ajax({
        url: '/api/me/tokens',
        method: 'GET',
        dataType: 'json',
        headers: PkReg.Auth.ajaxHeaders(),
        success: function(resp) {
          self.tokens = resp || [];
          self.render();
        },
        error: function(xhr) {
          if (xhr.status === 401) {
            PkReg.Auth.logout();
            Backbone.history.navigate('login', { trigger: true });
          }
        }
      });
    },

    onCreate: function(e) {
      e.preventDefault();
      var name = this.$('.token-name-input').val().trim();
      var ns = this.$('.token-namespace-input').val().trim();
      if (!name) return;

      var self = this;
      $.ajax({
        url: '/api/me/tokens',
        method: 'POST',
        contentType: 'application/json',
        headers: PkReg.Auth.ajaxHeaders(),
        data: JSON.stringify({ namespace: ns, name: name }),
        dataType: 'json',
        success: function(resp) {
          self.newToken = resp.raw_token;
          self.fetchTokens();
        },
        error: function(xhr) {
          var msg = 'Failed to create token';
          try { msg = JSON.parse(xhr.responseText).error; } catch(e) {}
          alert(msg);
        }
      });
    },

    onRevoke: function(e) {
      e.preventDefault();
      var tokenId = $(e.currentTarget).data('token-id');
      if (!confirm('Revoke this token? This cannot be undone.')) return;

      var self = this;
      $.ajax({
        url: '/api/me/tokens/' + encodeURIComponent(tokenId),
        method: 'DELETE',
        headers: PkReg.Auth.ajaxHeaders(),
        success: function() {
          self.fetchTokens();
        },
        error: function(xhr) {
          var msg = 'Failed to revoke token';
          try { msg = JSON.parse(xhr.responseText).error; } catch(e) {}
          alert(msg);
        }
      });
    }
  });

})();
