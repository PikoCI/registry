// models.js - Backbone Models
(function() {
  'use strict';

  PkReg.Models = {};

  PkReg.Models.RegType = Backbone.Model.extend({
    defaults: {
      ID: '',
      NamespaceID: '',
      NamespaceName: '',
      Name: '',
      Kind: '',
      Description: '',
      Repository: '',
      Homepage: '',
      License: '',
      Public: true,
      CreatedAt: ''
    }
  });

  PkReg.Models.Version = Backbone.Model.extend({
    defaults: {
      ID: '',
      TypeID: '',
      Version: '',
      Content: '',
      Readme: '',
      Params: null,
      Examples: null,
      Yanked: false,
      Deprecated: false,
      DeprecatedMessage: '',
      SuccessorVersion: '',
      Downloads: 0,
      CreatedAt: '',
      CreatedBy: ''
    }
  });

  PkReg.Models.Tag = Backbone.Model.extend({
    defaults: {
      Tag: '',
      Count: 0
    }
  });

  PkReg.Models.Token = Backbone.Model.extend({
    idAttribute: 'ID',
    defaults: {
      ID: '',
      NamespaceID: '',
      Name: '',
      Prefix: '',
      CreatedAt: '',
      ExpiresAt: null
    }
  });

  PkReg.Models.User = Backbone.Model.extend({
    defaults: {
      ID: '',
      GitHubID: '',
      Username: '',
      AvatarURL: '',
      CreatedAt: ''
    }
  });

  PkReg.Models.Namespace = Backbone.Model.extend({
    defaults: {
      ID: '',
      Name: '',
      Type: '',
      Public: true,
      CreatedAt: ''
    }
  });

})();
