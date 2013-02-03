/*
 *
 * From Exekiel Victor.  This fixes a pretty glaring issue with Angular's http
 * service, which is that it posts form data as application/json instead of
 * application/x-www-form-urlencoded.
 *
 * This is fine if you're working with an HTTP server that expects application/json.
 * But if you're working with one that just supports the actual W3C recommendation
 * for HTML form data, like, oh, Go's http package, the data never shows up in the
 * server - or rather, it shows up, as an unparsed string of JSON, as a form field
 * name.
 *
 * Here's a client-side solution.  It hooks into $httpProvider, setting the default
 * headers for posts to application/x-www-form-urlencoded (per the HTML rec) and
 * inserts a transformRequest function that serializes objects into name/value pairs.
 *
 * http://victorblog.com/2012/12/20/make-angularjs-http-service-behave-like-jquery-ajax/
 */
angular.module('httpFix', [], function($httpProvider)
{
  // Use x-www-form-urlencoded Content-Type
  $httpProvider.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded;charset=utf-8';

  // Override $http service's default transformRequest
  $httpProvider.defaults.transformRequest = [function(data)
  {
    /**
     * The workhorse; converts an object to x-www-form-urlencoded serialization.
     * @param {Object} obj
     * @return {String}
     */
    var param = function(obj)
    {
      var query = '';
      var name, value, fullSubName, subValue, innerObj, i;

      for(name in obj)
      {
        value = obj[name];

        if(value instanceof Array)
        {
          for(i=0; i<value.length; ++i)
          {
            subValue = value[i];
            fullSubName = name + '[' + i + ']';
            innerObj = {};
            innerObj[fullSubName] = subValue;
            query += param(innerObj) + '&';
          }
        }
        else if(value instanceof Object)
        {
          for(subName in value)
          {
            subValue = value[subName];
            fullSubName = name + '[' + subName + ']';
            innerObj = {};
            innerObj[fullSubName] = subValue;
            query += param(innerObj) + '&';
          }
        }
        else if(value !== undefined && value !== null)
        {
          query += encodeURIComponent(name) + '=' + encodeURIComponent(value) + '&';
        }
      }

      return query.length ? query.substr(0, query.length - 1) : query;
    };

    return angular.isObject(data) && String(data) !== '[object File]' ? param(data) : data;
  }];
});
