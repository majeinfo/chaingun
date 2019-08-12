"""server URL Configuration

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/2.0/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  path('', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  path('', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.urls import include, path
    2. Add a URL to urlpatterns:  path('blog/', include('blog.urls'))
"""
from django.contrib import admin
from django.urls import include, path
from . import views

urlpatterns = [
    #path('admin/', admin.site.urls),
    path('', views.index, name='home'),
    path('get_injectors', views.get_injectors, name='get_injectors'),
    path('clean_results', views.clean_results, name='clean_results'),
    path('store_results/<str:repository>/<str:resultname>/<str:scriptname>', views.store_results, name='store_results'),
    path('merge_results/<str:repository>', views.merge_results, name='merge_results'),
]
