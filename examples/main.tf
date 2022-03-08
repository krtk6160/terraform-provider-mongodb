terraform {
  required_version = ">= 0.13"

  required_providers {
    mongodb = {
      source = "registry.terraform.io/krtk6160/mongodb"
      version = "9.9.9"
    }
  }
}

locals {
  mongodb_database = "galoy"

  mongodb_views = [
    {
      "name" : "view1"
      "view_on" : "coll1"
      "pipeline" : [{
        "$project" : {
          "field1" : 1,
          "field2" : 1,
        }
      }]
    },
    {
      "name" : "view2"
      "view_on" : "coll2"
      "pipeline" : [{
        "$project" : {
          "field1" : 1,
          "field2" : 1,
        }
      }]
    },
    {
      "name" : "view3"
      "view_on" : "coll2"
      "pipeline" : [{
        "$project" : {
          "field1" : 1,
          "field2" : 1,
        }
      }]
    }
  ]
}

provider "mongodb" {
  host = "127.0.0.1"
  port = "27017"
  username = "root"
  password = "password"
  ssl = false
  auth_database = "admin"
}

resource "mongodb_db_role" "read_views" {
  name     = "data_read_views"
  database = local.mongodb_database

  dynamic "privilege" {
    for_each = local.mongodb_views
    content {
      db         = local.mongodb_database
      collection = privilege.value.name
      actions    = ["find"]
    }
  }
}

resource "mongodb_db_user" "user" {
  depends_on = [mongodb_db_role.read_views]

  auth_database = local.mongodb_database
  name          = "user"
  password      = "pass"
  role {
    role = mongodb_db_role.read_views.name
    db   = local.mongodb_database
  }
}

resource "mongodb_db_view" "view" {
  for_each = { for view in local.mongodb_views : view.name => view }

  name     = each.value.name
  database = local.mongodb_database
  view_on  = each.value.view_on
  pipeline = jsonencode(each.value.pipeline)
}
