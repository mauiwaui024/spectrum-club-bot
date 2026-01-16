export interface CalendarAPIResponse {
  view: string;
  current_date: string;
  start_date: string;
  end_date: string;
  prev_date: string;
  next_date: string;
  is_coach: boolean;
  user_name: string;
  user_id: string;
  week_days?: WeekDayHeader[];
  calendar_days?: CalendarDay[];
  week_days_data?: WeekDayData[];
  events?: CalendarEvent[];
  time_slots?: string[];
  training_days?: ScheduleDay[];
}

export interface WeekDayHeader {
  name: string;
  day: string;
}

export interface CalendarDay {
  date: string;
  is_today: boolean;
  is_other_month: boolean;
  events: CalendarEvent[];
}

export interface WeekDayData {
  name: string;
  day: number;
  date: string;
  is_today: boolean;
  events: CalendarEvent[];
}

export interface CalendarEvent {
  id: number;
  title: string;
  time: string;
  coach: string;
  top?: number;
  height?: number;
  color_index: number;
  user_id: string;
}

export interface ScheduleDay {
  date: string;
  trainings: TrainingView[];
}

export interface TrainingView {
  id: number;
  group_name: string;
  start_time: string;
  end_time: string;
  coach_name: string;
  participants: number;
  participant_names: string;
  max_participants: number;
  can_register: boolean;
  is_registered: boolean;
  is_full: boolean;
  color_index: number;
}

export interface TrainingDetails {
  training: {
    id: number;
    group_name: string;
    training_date: string;
    start_time: string;
    end_time: string;
    coach_name: string;
    description: string;
    max_participants: number | null;
  };
  participants: Participant[];
  participants_count: number;
  is_coach: boolean;
  is_training_coach: boolean;
  can_mark_attendance: boolean;
  is_registered: boolean;
  can_register: boolean;
  is_full: boolean;
  is_past: boolean;
  current_time: string;
}

export interface Participant {
  student_id: number;
  student_name: string;
  created_at: string;
}
