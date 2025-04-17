import React from "react";

const User = ({ user, compact = false }) => {
  if (!user) return null;

  // Generate initials for the avatar
  const getInitials = (name) => {
    return name
      .split(' ')
      .map(word => word[0])
      .join('')
      .toUpperCase()
      .substring(0, 2);
  };

  // Generate a consistent color based on user id
  const getAvatarColor = (id) => {
    const colors = [
      "bg-blue-500", "bg-green-500", "bg-yellow-500", 
      "bg-red-500", "bg-purple-500", "bg-pink-500", 
      "bg-indigo-500", "bg-teal-500"
    ];
    return colors[id % colors.length];
  };

  return (
    <div className={`flex items-center ${compact ? 'space-x-2 p-1.5' : 'space-x-3 p-2'}`}>
      {/* Avatar with user initials */}
      <div className={`flex-shrink-0 ${compact ? 'w-8 h-8' : 'w-10 h-10'} rounded-full flex items-center justify-center text-white font-bold ${getAvatarColor(user.id)}`}>
        {getInitials(user.name)}
      </div>
      
      {/* User details */}
      <div className="flex-grow overflow-hidden">
        <h3 className={`${compact ? 'text-sm' : 'text-md'} font-medium text-gray-900 truncate`}>
          {user.name}
        </h3>
        {(!compact || (compact && user.email)) && (
          <p className={`${compact ? 'text-xs' : 'text-sm'} text-gray-500 truncate`}>
            {user.email}
          </p>
        )}
      </div>
    </div>
  );
};

export default User;