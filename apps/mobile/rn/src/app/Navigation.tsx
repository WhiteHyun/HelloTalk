import React from 'react';
import {ActivityIndicator, View} from 'react-native';
import {NavigationContainer} from '@react-navigation/native';
import {createNativeStackNavigator} from '@react-navigation/native-stack';
import {createBottomTabNavigator} from '@react-navigation/bottom-tabs';

import {useAuth} from './AuthContext';
import {LoginScreen} from '../features/auth/screens/LoginScreen';
import {SignupScreen} from '../features/auth/screens/SignupScreen';
import {FeedScreen} from '../features/feed/screens/FeedScreen';
import {CreatePostScreen} from '../features/feed/screens/CreatePostScreen';
import {ProfileScreen} from '../features/profile/screens/ProfileScreen';

const AuthStack = createNativeStackNavigator();
const MainTab = createBottomTabNavigator();
const FeedStack = createNativeStackNavigator();

function FeedStackScreen() {
  return (
    <FeedStack.Navigator>
      <FeedStack.Screen
        name="FeedHome"
        component={FeedScreen}
        options={{title: '피드'}}
      />
      <FeedStack.Screen
        name="CreatePost"
        component={CreatePostScreen}
        options={{title: '게시글 작성'}}
      />
    </FeedStack.Navigator>
  );
}

function MainTabScreen() {
  return (
    <MainTab.Navigator
      screenOptions={{
        tabBarActiveTintColor: '#007AFF',
        tabBarInactiveTintColor: '#999',
      }}>
      <MainTab.Screen
        name="Feed"
        component={FeedStackScreen}
        options={{
          headerShown: false,
          tabBarLabel: '피드',
        }}
      />
      <MainTab.Screen
        name="Profile"
        component={ProfileScreen}
        options={{
          title: '프로필',
          tabBarLabel: '프로필',
        }}
      />
    </MainTab.Navigator>
  );
}

function AuthStackScreen() {
  return (
    <AuthStack.Navigator screenOptions={{headerShown: false}}>
      <AuthStack.Screen name="Login" component={LoginScreen} />
      <AuthStack.Screen name="Signup" component={SignupScreen} />
    </AuthStack.Navigator>
  );
}

export function Navigation() {
  const {isLoading, user} = useAuth();

  if (isLoading) {
    return (
      <View style={{flex: 1, justifyContent: 'center', alignItems: 'center'}}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  return (
    <NavigationContainer>
      {user ? <MainTabScreen /> : <AuthStackScreen />}
    </NavigationContainer>
  );
}
