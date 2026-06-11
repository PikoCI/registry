// collections.js - Backbone Collections
(function() {
  'use strict';

  PkReg.Collections = {};

  PkReg.Collections.SearchResults = Backbone.Collection.extend({
    model: PkReg.Models.RegType,
    url: '/api/plugins',
    totalCount: 0,

    parse: function(response) {
      this.totalCount = response.TotalCount || 0;
      return response.Types || [];
    }
  });

  PkReg.Collections.Versions = Backbone.Collection.extend({
    model: PkReg.Models.Version,
    comparator: function(v) {
      return -new Date(v.get('CreatedAt')).getTime();
    }
  });

  PkReg.Collections.Tags = Backbone.Collection.extend({
    model: PkReg.Models.Tag,
    url: '/api/tags',
    comparator: function(t) {
      return -t.get('Count');
    }
  });

  PkReg.Collections.MyTypes = Backbone.Collection.extend({
    model: PkReg.Models.RegType,
    url: '/api/me/types',

    sync: function(method, collection, options) {
      options = options || {};
      options.headers = PkReg.Auth.ajaxHeaders();
      return Backbone.sync.call(this, method, collection, options);
    }
  });

  PkReg.Collections.MyTokens = Backbone.Collection.extend({
    model: PkReg.Models.Token,
    url: '/api/me/tokens',

    sync: function(method, collection, options) {
      options = options || {};
      options.headers = PkReg.Auth.ajaxHeaders();
      return Backbone.sync.call(this, method, collection, options);
    }
  });

})();
